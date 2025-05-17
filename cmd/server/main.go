package main

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
	"github.com/zero-shubham/surveysvc/api"
)

const (
	MaxIdelConn = 5
	DbUrlEnv    = "DATABASE_URL"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx := context.Background()

	// tp, err := config.Init()
	// if err != nil {
	// 	log.Fatal().Err(err).Send()
	// }
	// defer func() {
	// 	if err := tp.Shutdown(ctx); err != nil {
	// 		log.Err(err).Msg("error shutting down tracer provider")
	// 	}
	// }()

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

	api.NewRouter(&log.Logger, dbConn).Start(ctx)
	<-ctx.Done()

}
