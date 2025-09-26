package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/SirWaithaka/payments-api/pkg/logger"
	"github.com/SirWaithaka/payments-api/src/api/rest"
	"github.com/SirWaithaka/payments-api/src/config"
	dipkg "github.com/SirWaithaka/payments-api/src/di"
	"github.com/SirWaithaka/payments-api/src/events/listener"
	"github.com/SirWaithaka/payments-api/src/events/publisher"
	"github.com/SirWaithaka/payments-api/src/storage"
)

func main() {
	// load application configs
	var cfg config.Config
	if err := config.FromEnv(&cfg); err != nil {
		panic(errors.Wrap(err, "env configs could not be loaded"))
	}

	// create context that listens to on sigterm and sigint
	mCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	// set default logger
	l := logger.New(&logger.Config{LogMode: cfg.LogLevel, Service: cfg.ServiceName})
	logger.SetDefaultLogger(l)
	zerolog.DefaultContextLogger = &l
	// add logger to context
	mCtx = l.WithContext(mCtx)

	// create db connection
	db, err := storage.NewDatabase(cfg)
	if err != nil {
		l.WithLevel(zerolog.FatalLevel).Err(err).Msg("could not connect to db")
		return
	}
	l.Info().Msg("db connection succeeded")

	// create an instance of publisher
	pub := publisher.New(cfg.Kafka)

	// initialize DI container
	di := dipkg.New(cfg, db, pub)

	// create an error group
	g, gCtx := errgroup.WithContext(mCtx)

	// create an instance of http rest server
	app := rest.New(gCtx, di)
	g.Go(app.Start)
	g.Go(app.Stop)

	// create an instance of listener and start
	ln := listener.New(gCtx, cfg.Kafka, di)
	g.Go(ln.Listen)
	g.Go(ln.Close)

	// wait for all goroutines in a g group
	if err = g.Wait(); err != nil {
		return
	}
	l.Info().Msg("main shutting down")

	// close publisher
	if err = pub.Close(); err != nil {
		l.WithLevel(zerolog.FatalLevel).Err(err).Msg("publisher close error")
	}

	// shutdown db
	db.Close()

	l.Info().Msg("main shut down")

}
