package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/di"
	"github.com/SirWaithaka/payments-api/internal/pkg/http/middlewares"
	"github.com/SirWaithaka/payments-api/internal/pkg/http/middlewares/ginzerolog"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

func New(c context.Context, di *di.DI) *Server {
	l := logger.New(&logger.Config{})

	return &Server{
		logger: l,
		ctx:    c,
		server: newServer(di.Cfg.HTTPPort, &l),
	}
}

type Server struct {
	server *http.Server
	ctx    context.Context
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

func newServer(port string, logger *zerolog.Logger) *http.Server {
	engine := gin.New()
	gin.SetMode(gin.ReleaseMode)

	// add middlewares to server
	engine.Use(gin.Recovery())
	engine.Use(ginzerolog.New(ginzerolog.Config{Logger: logger}))
	engine.Use(middlewares.ErrorHandler())

	routes(engine)

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: engine.Handler(),
	}
}
