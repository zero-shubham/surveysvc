package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/exaring/otelpgx"
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
	OtelCollectorEnv      = "OLTP_HTTP_ENDPOINT"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	tp, mp, err := config.Init(ctx, os.Getenv(OtelCollectorEnv), config.ServiceName+"-consumer")
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Err(err).Msg("error shutting down tracer provider")
		}

		if err := mp.Shutdown(ctx); err != nil {
			log.Err(err).Msg("error shutting down metric provider")
		}

		log.Info().Msg("shutdown of tracer and meter complete")
	}()

	DB_URL := os.Getenv(DbUrlEnv)
	dbConfig, err := pgxpool.ParseConfig(DB_URL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create a config")
	}

	dbConfig.ConnConfig.Tracer = otelpgx.NewTracer(otelpgx.WithTracerProvider(tp))
	dbConfig.MaxConns = MaxIdelConn
	dbConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	dbConn, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to db conn")
	}

	err = otelpgx.RecordStats(dbConn, otelpgx.WithStatsMeterProvider(mp))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to add metrics to db")
	}

	svc := internal.NewService(config.GetLogger(), dbConn)
	consumer := messaging.NewKafkaConsumer(
		[]string{os.Getenv(KafkaBrokerEnv)},
		os.Getenv(KafkaTopicConsumeEnv),
		os.Getenv(KafkaConsumerGroupEnv)+os.Getenv(messaging.PodNameEnv),
		svc.HandleAnswer,
		os.Getenv(KafkaDeadLetterEnv),
		config.GetLogger(),
		tp,
	)
	consumer.Start(ctx, 2, mp)

	<-ctx.Done()

}
