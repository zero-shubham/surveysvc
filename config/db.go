package config

import (
	"database/sql"
	"os"

	"github.com/golang-migrate/migrate/v4"
	driver "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	DbUrlEnv = "DATABASE_URL"
)

func MigrateUp() error {
	DB_URL := os.Getenv(DbUrlEnv)

	db, err := sql.Open("pgx/v5", DB_URL)
	if err != nil {
		return err
	}

	dr, err := driver.WithInstance(db, &driver.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file:///authsvc/db/migrations",
		"pgx/v5", dr)
	if err != nil {
		return err
	}
	return m.Up()
}

func MigrateDown() error {
	DB_URL := os.Getenv(DbUrlEnv)

	db, err := sql.Open("pgx/v5", DB_URL)
	if err != nil {
		return err
	}

	dr, err := driver.WithInstance(db, &driver.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file:///authsvc/db/migrations",
		"pgx/v5", dr)
	if err != nil {
		return err
	}

	return m.Drop()
}
