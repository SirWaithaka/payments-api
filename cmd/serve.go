package cmd

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/SirWaithaka/payments-api/pkg/graceful"
	"github.com/SirWaithaka/payments-api/src/api/rest"
	"github.com/SirWaithaka/payments-api/src/config"
	dipkg "github.com/SirWaithaka/payments-api/src/di"
	"github.com/SirWaithaka/payments-api/src/events/listener"
	"github.com/SirWaithaka/payments-api/src/events/publisher"
	"github.com/SirWaithaka/payments-api/src/storage"
)

func serveApi(ctx context.Context, cfg config.Config) (func() error, error) {

	// create db connection
	db, err := storage.NewDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// create a publisher instance
	pub := publisher.New(cfg.Kafka)

	// create DI container
	di := dipkg.New(cfg, db, pub)

	// instance of the rest server
	server := rest.New(di)

	return func() error {
		return graceful.GracefulContext(ctx, server.Start, server.Stop)
	}, nil

}

func ServeAll() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		g, ctx := errgroup.WithContext(ctx)
		cmd.SetContext(ctx)

		// load application configs
		var cfg config.Config
		if err := config.FromEnv(&cfg); err != nil {
			return errors.Wrap(err, "env configs could not be loaded")
		}

		server, err := serveApi(ctx, cfg)
		if err != nil {
			return err
		}
		ln := listener.New(ctx, cfg.Kafka)
	}
}

func NewServeCmd() *cobra.Command {
	// cmd represents the serve command
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the payments server",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	return cmd
}
