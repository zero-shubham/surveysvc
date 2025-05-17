package main

// import (
// 	"context"
// 	"os"
// 	"os/signal"

// 	"github.com/jackc/pgx/v5"
// 	"github.com/rs/zerolog"
// 	"github.com/rs/zerolog/log"
// 	db "github.com/zero-shubham/authsvc/db/orm"
// 	password "github.com/zero-shubham/authsvc/internal"
// )

// const (
// 	MaxIdelConn        = 5
// 	DbUrlEnv           = "DATABASE_URL"
// 	SuperAdminEmailEnv = "SUPERADMIN_EMAIL"
// 	SuperAdminPassEnv  = "SUPERADMIN_PASS"
// )

// func main() {
// 	zerolog.SetGlobalLevel(zerolog.InfoLevel)

// 	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
// 	defer stop()

// 	DB_URL := os.Getenv(DbUrlEnv)

// 	superAdminEmail := os.Getenv(SuperAdminEmailEnv)
// 	superAdminPass := os.Getenv(SuperAdminPassEnv)

// 	if superAdminEmail == "" || superAdminPass == "" {
// 		log.Fatal().Msgf("missing %s or %s env", SuperAdminEmailEnv, SuperAdminPassEnv)
// 	}
// 	log.Info().Str("superadmin_pass", superAdminPass).Str("superadmin_email", superAdminEmail).Msg("ready to seed")

// 	dbConn, err := pgx.Connect(ctx, DB_URL)
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("Failed to create a config")
// 	}

// 	defer func() {
// 		err = dbConn.Close(ctx)
// 		if err != nil {
// 			log.Fatal().Err(err).Msg("failed to close db conn")
// 		}
// 	}()

// 	orm := db.New(dbConn)

// 	found, err := orm.GetUserAndAppGrpByEmail(ctx, superAdminEmail)
// 	if err != nil && err != pgx.ErrNoRows {
// 		log.Fatal().Err(err).Msg("failed to seed")
// 	}
// 	if found.Email == superAdminEmail {
// 		log.Info().Msg("db already seeded")
// 		return
// 	}

// 	appGrp, err := orm.CreateAppGroup(ctx, db.CreateAppGroupParams{
// 		Name:   "superadmin",
// 		Scopes: []string{"*:*"},
// 	})
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("failed to seed app group")
// 	}

// 	hashedPass, err := password.HashPassword(superAdminPass)
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("failed to hash password")
// 	}

// 	user, err := orm.CreateUser(ctx, db.CreateUserParams{
// 		Email:      superAdminEmail,
// 		Password:   hashedPass,
// 		AppGroupID: appGrp.ID,
// 	})
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("failed to seed user")
// 	}
// 	log.Info().Str("user_id", user.ID.String()).Msg("successfully seeded db")

// }
