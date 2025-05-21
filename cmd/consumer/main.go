package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
	"github.com/zero-shubham/surveysvc/config"
	"github.com/zero-shubham/surveysvc/internal"
	"github.com/zero-shubham/surveysvc/transport/messaging"
)

const (
	MaxIdelConn           = 5
	DbUrlEnv              = "DATABASE_URL"
	KafkaBrokerEnv        = "KAFKA_BROKER_URI"
	KafkaTopicConsumeEnv  = "KAFKA_CONSUMER_TOPIC"
	KafkaDeadLetterEnv    = "KAFKA_DEADLETTER_TOPIC"
	KafkaConsumerGroupEnv = "KAFKA_CONSUMER_GROUP"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	tp, err := config.Init()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Err(err).Msg("error shutting down tracer provider")
		}
	}()

	DB_URL := os.Getenv(DbUrlEnv)
	dbConfig, err := pgxpool.ParseConfig(DB_URL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create a config")
	}
	dbConfig.MaxConns = MaxIdelConn
	dbConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	dbConn, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create a config")
	}

	svc := internal.NewService(&log.Logger, dbConn)
	consumer := messaging.NewKafkaConsumer(
		[]string{os.Getenv(KafkaBrokerEnv)},
		os.Getenv(KafkaTopicConsumeEnv),
		os.Getenv(KafkaConsumerGroupEnv),
		svc.HandleAnswer,
		os.Getenv(KafkaDeadLetterEnv),
		&log.Logger,
		tp,
	)
	consumer.Start(ctx, 2)

	<-ctx.Done()

}
