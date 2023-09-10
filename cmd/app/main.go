package main

import (
	"context"
	"os"
	"time"

	"github.com/Salam4nder/user/internal/config"
	"github.com/Salam4nder/user/internal/db"
	"github.com/Salam4nder/user/internal/grpc"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stimtech/go-migration"
	"go.uber.org/zap"
)

const (
	//PingTimeout is the maximum duration for waiting on ping.
	PingTimeout = 5 * time.Second
	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	ReadTimeout = 10 * time.Second
	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read.
	WriteTimeout = 10 * time.Second
	// EnvironmentDev is the development environment.
	EnvironmentDev = "dev"
)

func main() {
	cfg, err := config.New()
	fatalExitOnErr(err)

	var logger zerolog.Logger

	if cfg.Environment == "dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		fatalExitOnErr(err)

		defer file.Close()

		log.Logger = log.Output(file)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		PingTimeout,
	)
	defer cancel()

	sql, err := db.NewSQLDatabase(ctx, cfg.PSQL)
	fatalExitOnErr(err)

	log.Info().Msg("successfully connected to database...")

	migration := migration.New(sql.DB(), zap.NewNop())

	if err := migration.Migrate(); err != nil {
		fatalExitOnErr(err)
	}
	log.Info().Msg("successfully migrated database...")

	service, err := grpc.NewUserService(sql, cfg.Service)
	fatalExitOnErr(err)

	server := grpc.NewServer(service, &cfg.Server, &logger)

	go server.ServeGRPCGateway()

	err = server.ServeGRPC()
	fatalExitOnErr(err)
}

func fatalExitOnErr(err error) {
	if err != nil {
		log.Fatal().Err(err).Msg("fatal exit: failed to start user service")
	}
}
