package messaging

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
	"github.com/zero-shubham/surveysvc/config"
	"github.com/zero-shubham/surveysvc/transport/tcp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const PodNameEnv = "POD_NAME"

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
	tp *sdktrace.TracerProvider,
) *KafkaConsumer {
	logger.Info().Str("consumer_group", consumerGroupID).Msg("instantiating new consumer")

	tracer := tp.Tracer(config.ServiceName + "-consumer-" + os.Getenv(PodNameEnv))

	dialer := tcp.NewInstrumentedDialer(time.Second*30, time.Minute*60, tracer, logger)

	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  consumerGroupID,
			MaxBytes: 10e6,
			Dialer:   dialer.Dialer,
		}),
		handler: handler,
		deadletter: kafka.NewWriter(kafka.WriterConfig{
			Brokers:  brokers,
			Topic:    deadletterTopic,
			Balancer: &kafka.LeastBytes{},
			Dialer:   dialer.Dialer,
		}),
		logger: logger,
		topic:  topic,
		trace:  tracer,
	}

}

func (kh *KafkaConsumer) Start(ctx context.Context, workerCount int, mp metric.MeterProvider) {

	meter := mp.Meter(config.ServiceName + "-consumer-" + os.Getenv(PodNameEnv))

	kh.messageChan = make(chan *kafka.Message, workerCount)
	for i := 0; i < workerCount; i++ {
		go kh.workerStart(ctx, i)
	}

	msgCounter, err := meter.Int64Counter(
		"message_counter",
		metric.WithDescription("Counts the  total messages read"),
		metric.WithUnit("1"),
	)
	if err != nil {
		kh.logger.Fatal().Err(err).Msg("failed to instantiate msg counter")
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
				kh.logger.Info().Msg("fetching message")
				ctx, span := kh.trace.Start(ctx, "fetch-answer")

				m, err := kh.reader.FetchMessage(ctx)
				if err != nil {
					kh.logger.Err(err).Msg("failed to fetch message")
					continue
				}

				kh.logger.Info().Msg("fetched message")

				otelAttrs := make([]attribute.KeyValue, len(m.Headers))
				for _, header := range m.Headers {
					otelAttrs = append(otelAttrs, attribute.KeyValue{
						Key:   attribute.Key(header.Key),
						Value: attribute.StringValue(string(header.Value)),
					})
				}
				msgCounter.Add(ctx, 1, metric.WithAttributes(attribute.Float64("timestamp", float64(time.Now().Unix()))))

				span.SetAttributes(otelAttrs...)

				kh.messageChan <- &m
				span.End()
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

func (kh *KafkaConsumer) workerStart(ctx context.Context, id int) {
	kh.logger.Info().Msg("starting consumer worker")
	for {
		select {
		case <-ctx.Done():
			kh.logger.Info().Msg("shutting down consumer worker")
			return

		case m := <-kh.messageChan:
			ctx, span := kh.trace.Start(ctx, "process-answer")
			otelAttrs := make([]attribute.KeyValue, len(m.Headers))
			for _, header := range m.Headers {
				otelAttrs = append(otelAttrs, attribute.KeyValue{
					Key:   attribute.Key(header.Key),
					Value: attribute.StringValue(string(header.Value)),
				})
			}
			span.SetAttributes(otelAttrs...)

			kh.logger.Info().Msgf("message received on worker %d", id)

			err := ExponentialRetry(3, func() error {
				err := kh.handler(ctx, m)
				if err != nil {
					kh.logger.Err(err).Msg("failure from message handler, retrying..")
				}
				return err
			})
			if err != nil {
				kh.logger.Err(err).Msg("failed to process message")
				ctx, span := kh.trace.Start(ctx, "dead-answer")

				err = kh.deadletter.WriteMessages(ctx, *m)
				if err != nil {
					kh.logger.Err(err).Msg("error while wrtiging to dead leader")
				}

				span.End()
			}

			ctx, commitSpan := kh.trace.Start(ctx, "commit-answer")
			if err := kh.reader.CommitMessages(ctx, *m); err != nil {
				kh.logger.Err(err).Msg("failed to commit messages")
			}

			commitSpan.End()
			span.End()
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
