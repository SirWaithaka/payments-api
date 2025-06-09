package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/SirWaithaka/payments-api/internal/api/rest"
	"github.com/SirWaithaka/payments-api/internal/config"
	dipkg "github.com/SirWaithaka/payments-api/internal/di"
)

func main() {
	// load application configs
	var cfg config.Config
	if err := config.FromEnv(&cfg); err != nil {
		panic(errors.Wrap(err, "env configs could not be loaded"))
	}

	// create context that listens on sigterm and sigint
	mCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	// initialize DI container
	di := dipkg.New(cfg)

	// create error group
	g, gCtx := errgroup.WithContext(mCtx)

	// create instance of http rest server
	app := rest.New(gCtx, di)
	g.Go(app.Start)
	g.Go(app.Stop)

	// wait for all goroutines in g group
	if err := g.Wait(); err != nil {
		return
	}
}
