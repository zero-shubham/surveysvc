package messaging

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

type HandlerFunc func(*kafka.Message) error
type KafkaHandler struct {
	readers    []*kafka.Reader
	handler    HandlerFunc
	deadletter *kafka.Writer
	logger     *zerolog.Logger
	topic      string
}

func NewKafkaHandler(
	brokers []string,
	topic string,
	consumerGroupIDs []string,
	handler HandlerFunc,
	deadletterTopic string,
	logger *zerolog.Logger,
) *KafkaHandler {
	readers := make([]*kafka.Reader, len(consumerGroupIDs))

	for _, groupID := range consumerGroupIDs {
		readers = append(readers, kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MaxBytes: 10e6,
		}))
	}

	return &KafkaHandler{
		readers: readers,
		handler: handler,
		deadletter: kafka.NewWriter(kafka.WriterConfig{
			Brokers:  brokers,
			Topic:    deadletterTopic,
			Balancer: &kafka.LeastBytes{},
		}),
		logger: logger,
		topic:  topic,
	}
}

func (kh *KafkaHandler) Start(ctx context.Context) {

	for id := range kh.readers {
		go kh.WorkerStart(ctx, id)
	}
}

func (kh *KafkaHandler) Stop(stop context.CancelFunc) {
	err := kh.deadletter.Close()
	if err != nil {
		kh.logger.Err(err).Msg("failed to close dead letter")
	}

	for _, reader := range kh.readers {
		err := reader.Close()
		if err != nil {
			kh.logger.Err(err).Msg("failed to close reader")
		}
	}
	kh.logger.Info().Msg("stopped kafka handler")
}

func (kh *KafkaHandler) WorkerStart(ctx context.Context, id int) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer kh.Stop(stop)

	kh.logger.Info().Str("topic", kh.topic).Int("worker", id).Msg("starting worker consumer")
	for {
		select {
		case <-ctx.Done():
			kh.logger.Info().Str("topic", kh.topic).Msg("stopping kafka handler")
			break

		default:
			m, err := kh.readers[id].FetchMessage(ctx)
			if err != nil {
				kh.logger.Err(err).Msg("failed to fetch messages")
				break
			}

			err = ExponentialRetry(3, func() error {
				return kh.handler(&m)
			})
			if err != nil {
				err = kh.deadletter.WriteMessages(ctx, m)
				if err != nil {
					kh.logger.Err(err).Msg("error while wrtiging to dead leader")
				}
			}

			if err := kh.readers[id].CommitMessages(ctx, m); err != nil {
				log.Fatal("failed to commit messages:", err)
			}

		}
	}
}

func ExponentialRetry(maxRetry int, execute func() error) error {
	counter := 1
	var err error
	for err = execute(); err != nil && counter < maxRetry; {
		time.Sleep(time.Second * (time.Duration(counter + (2 * counter))))
		counter++
	}
	return err
}
