package main

import (
	"context"
	"os"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
	"github.com/zero-shubham/surveysvc/api"
	"github.com/zero-shubham/surveysvc/config"
	"go.opentelemetry.io/otel"
)

const (
	MaxIdelConn      = 5
	OtelCollectorEnv = "OLTP_HTTP_ENDPOINT"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx := context.Background()

	tp, mp, err := config.Init(ctx, os.Getenv(OtelCollectorEnv), config.ServiceName)
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

	DB_URL := os.Getenv(config.DbUrlEnv)
	dbConfig, err := pgxpool.ParseConfig(DB_URL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create a config")
	}

	dbConfig.ConnConfig.Tracer = otelpgx.NewTracer(otelpgx.WithTracerProvider(otel.GetTracerProvider()))
	dbConfig.MaxConns = MaxIdelConn
	dbConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	dbConn, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to db conn")
	}

	err = otelpgx.RecordStats(dbConn)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to add metrics to db")
	}

	api.NewRouter(&log.Logger, dbConn).Start(ctx, tp, mp)
	<-ctx.Done()

}
