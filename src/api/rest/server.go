package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/pkg/http/middlewares"
	"github.com/SirWaithaka/payments-api/pkg/http/middlewares/ginzerolog"
	dipkg "github.com/SirWaithaka/payments-api/src/di"
)

func New(c context.Context, di *dipkg.DI) *Server {
	l := di.Cfg.Logger()

	engine := gin.New()
	gin.SetMode(gin.ReleaseMode)

	// add middlewares to server
	engine.Use(gin.Recovery())
	engine.Use(ginzerolog.New(ginzerolog.Config{Logger: l}))
	engine.Use(middlewares.ErrorHandler())
	// add health check route
	engine.GET("/health", middlewares.Healthcheck)

	routes(engine, di)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", di.Cfg.HTTPPort),
		Handler: engine.Handler(),
	}

	return &Server{
		ctx:    c,
		logger: *l,
		server: server,
	}
}

type Server struct {
	ctx    context.Context
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
func (server *Server) Stop() error {
	defer server.logger.Info().Msg("server stopped")

	<-server.ctx.Done()
	server.logger.Info().Msg("stopping http server ...")
	return server.server.Shutdown(server.ctx)
}
