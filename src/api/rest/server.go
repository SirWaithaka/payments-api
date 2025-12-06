package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	middlewares2 "github.com/SirWaithaka/payments-api/pkg/http/middlewares"
	ginzerolog2 "github.com/SirWaithaka/payments-api/pkg/http/middlewares/ginzerolog"
	"github.com/SirWaithaka/payments-api/pkg/logger"
	dipkg "github.com/SirWaithaka/payments-api/src/di"
)

func New(di *dipkg.DI) *Server {
	l := logger.New(&logger.Config{})

	engine := gin.New()
	gin.SetMode(gin.ReleaseMode)

	// add middlewares to server
	engine.Use(gin.Recovery())
	engine.Use(ginzerolog2.New(ginzerolog2.Config{Logger: &l}))
	engine.Use(middlewares2.ErrorHandler())
	// add health check route
	engine.GET("/health", middlewares2.Healthcheck)

	routes(engine, di)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", di.Cfg.HTTPPort),
		Handler: engine.Handler(),
	}

	return &Server{
		logger: l,
		server: server,
	}
}

type Server struct {
	server *http.Server
	logger zerolog.Logger
}

func (server *Server) Start() error {
	server.logger.Info().Msg("starting http server ...")

	if err := server.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		server.logger.WithLevel(zerolog.FatalLevel).Err(err).Msg("failed to start http server")
		return err
	}

	return nil
}

// Stop listens for an os signal and shuts down the server
func (server *Server) Stop(ctx context.Context) error {
	defer server.logger.Info().Msg("server stopped")

	<-ctx.Done()
	server.logger.Info().Msg("stopping http server ...")
	return server.server.Shutdown(ctx)
}
