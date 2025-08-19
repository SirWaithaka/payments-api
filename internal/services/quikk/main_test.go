package quikk_test

import (
	"os"
	"testing"

	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
	"github.com/SirWaithaka/payments-api/internal/testdata"
)

var inf *testdata.Infrastructure

func TestMain(m *testing.M) {
	cfg, err := testdata.LoadConfig()
	if err != nil {
		os.Exit(1)
	}

	l := logger.New(&logger.Config{LogMode: cfg.LogLevel})
	zerolog.DefaultContextLogger = &l

	// run setup
	inf, err = testdata.Setup(cfg)
	if err != nil {
		l.Fatal().Err(err).Msg("error setting up test infrastructure")
		os.Exit(1)
	}

	// run tests
	code := m.Run()

	// do some cleanup
	testdata.CleanUp(inf)

	os.Exit(code)
}
