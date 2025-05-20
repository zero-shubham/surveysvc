package messaging

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
	"github.com/zero-shubham/surveysvc/transport/tcp"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type HandlerFunc func(context.Context, *kafka.Message) error
type KafkaConsumer struct {
	reader      *kafka.Reader
	handler     HandlerFunc
	deadletter  *kafka.Writer
	logger      *zerolog.Logger
	topic       string
	messageChan chan *kafka.Message
	trace       trace.Tracer
}

func NewKafkaConsumer(
	brokers []string,
	topic string,
	consumerGroupID string,
	handler HandlerFunc,
	deadletterTopic string,
	logger *zerolog.Logger,
	workerCount int,
	tp *sdktrace.TracerProvider,
) *KafkaConsumer {
	tracer := tp.Tracer("surveysvc-consumer")
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  consumerGroupID,
			MaxBytes: 10e6,
		}),
		handler: handler,
		deadletter: kafka.NewWriter(kafka.WriterConfig{
			Brokers:  brokers,
			Topic:    deadletterTopic,
			Balancer: &kafka.LeastBytes{},
			Dialer:   tcp.NewInstrumentedDialer(time.Second*30, time.Minute*60, tracer).Dialer,
		}),
		logger:      logger,
		topic:       topic,
		messageChan: make(chan *kafka.Message, workerCount),
		trace:       tracer,
	}

}

func (kh *KafkaConsumer) Start(ctx context.Context) {

	for i := 0; i < len(kh.messageChan); i++ {
		go kh.workerStart(ctx)
	}

	go func() {
		kh.logger.Info().Str("topic", kh.topic).Msg("starting consumer")
		for {
			select {
			case <-ctx.Done():
				kh.logger.Info().Str("topic", kh.topic).Msg("stopping kafka handler")
				kh.Stop()
				return

			default:
				ctx, span := kh.trace.Start(ctx, "fetch-answers")
				m, err := kh.reader.FetchMessage(ctx)
				if err != nil {
					kh.logger.Err(err).Msg("failed to fetch messages")
					continue
				}

				for _, header := range m.Headers {
					span.SetAttributes(attribute.KeyValue{
						Key:   attribute.Key(header.Key),
						Value: attribute.StringValue(string(header.Value)),
					})
				}

				kh.messageChan <- &m
			}
		}
	}()
}

func (kh *KafkaConsumer) Stop() {
	err := kh.deadletter.Close()
	if err != nil {
		kh.logger.Err(err).Msg("failed to close dead letter")
	}

	err = kh.reader.Close()
	if err != nil {
		kh.logger.Err(err).Msg("failed to close reader")
	}
	kh.logger.Info().Msg("stopped kafka handler")
}

func (kh *KafkaConsumer) workerStart(ctx context.Context) {
	kh.logger.Info().Msg("starting consumer worker")
	for {
		select {
		case <-ctx.Done():
			kh.logger.Info().Msg("shutting down consumer worker")
			return

		case m := <-kh.messageChan:
			ctx, _ := kh.trace.Start(ctx, "process-answer")
			err := ExponentialRetry(3, func() error {
				return kh.handler(ctx, m)
			})
			if err != nil {
				ctx, _ := kh.trace.Start(ctx, "dead-answer")
				err = kh.deadletter.WriteMessages(ctx, *m)
				if err != nil {
					kh.logger.Err(err).Msg("error while wrtiging to dead leader")
				}
			}

			ctx, _ = kh.trace.Start(ctx, "commit-answer")
			if err := kh.reader.CommitMessages(ctx, *m); err != nil {
				kh.logger.Err(err).Msg("failed to commit messages")
			}
		}
	}

}

func ExponentialRetry(maxRetry int, execute func() error) error {
	counter := 1
	var err error
	for err = execute(); err != nil && counter < maxRetry; {
		time.Sleep(time.Second * (time.Duration(counter + (2 + counter))))
		counter++
	}
	return err
}
