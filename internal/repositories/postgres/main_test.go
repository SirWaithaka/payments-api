package postgres_test

import (
	"os"
	"reflect"
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

func AssertNil(t *testing.T, actual any) {
	if actual == nil {
		t.Errorf("expected nil, got %v", actual)
	}

	value := reflect.ValueOf(actual)
	switch value.Kind() {
	case
		reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.Slice, reflect.UnsafePointer:

		if !value.IsNil() {
			// For pointers, try to get the underlying value
			if value.Kind() == reflect.Ptr && value.Elem().IsValid() {
				t.Errorf("expected nil, got %v", value.Elem().Interface())
			} else {
				t.Errorf("expected nil, got %v", actual)
			}
		}

	default:
		return
	}

}
