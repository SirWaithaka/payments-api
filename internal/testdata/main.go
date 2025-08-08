package testdata

import (
	"fmt"
	"os"
	"sync"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"

	"github.com/SirWaithaka/payments-api/internal/config"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
	"github.com/SirWaithaka/payments-api/internal/repositories/postgres"
	"github.com/SirWaithaka/payments-api/internal/storage"
)

var (
	cfg *config.Config

	once sync.Once
)

type Infrastructure struct {
	Conf    *config.Config
	Storage *storage.Database
	pg      *gorm.DB
}

// GetSetEnv only sets env variable if not available
func GetSetEnv(k, v string) {
	if os.Getenv(k) == "" {
		_ = os.Setenv(k, v)
	}
}

func loadConfig() (config.Config, error) {
	GetSetEnv("POSTGRES_USER", "tester")
	GetSetEnv("POSTGRES_PASSWORD", "testingjkl")
	GetSetEnv("POSTGRES_HOST", "localhost")
	GetSetEnv("POSTGRES_PORT", "5432")
	GetSetEnv("POSTGRES_DATABASE", "test")
	GetSetEnv("POSTGRES_SCHEMA", "public")
	GetSetEnv("KAFKA_BROKERS", "url://fake-host")

	var conf config.Config
	err := config.FromEnv(&conf)
	if err != nil {
		return config.Config{}, err
	}

	return conf, nil
}

func LoadConfig() (*config.Config, error) {
	once.Do(func() {
		if cfg == nil {
			c, err := loadConfig()
			if err != nil {
				panic(err)
			}
			cfg = &c
		}
	})

	return cfg, nil
}

func Setup(cfg *config.Config) (*Infrastructure, error) {
	defer func() {
		if r := recover(); r != nil {
			if cfg != nil {
			}
		}
	}()

	conf := cfg
	if conf == nil {
		var err error
		conf, err = LoadConfig()
		if err != nil {
			return nil, err
		}
	}

	l := logger.New(&logger.Config{LogMode: conf.LogLevel})
	l.Info().Interface(logger.LData, conf).Msg("test configs loaded")

	// create a connection to the postgres host
	pg, err := storage.NewPostgresClient(conf.Postgres)
	if err != nil {
		return nil, err
	}
	l.Info().Msg("connected to db")

	// create new database to run tests
	conf.Postgres.DbName = "_" + ulid.Make().String()
	sql := fmt.Sprintf(`CREATE DATABASE "%v"`, conf.Postgres.DbName)
	l.Info().Msg(sql)
	result := pg.Exec(sql)
	if err = result.Error; err != nil {
		return nil, err
	}

	// creates connection to new test db
	store, err := storage.NewDatabase(*conf)
	if err != nil {
		return nil, err
	}

	// create db schema
	sql = fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS "%v"`, conf.Postgres.Schema)
	l.Info().Msg(sql)
	result = store.PG.Exec(sql)
	if err = result.Error; err != nil {
		return nil, err
	}

	// migrate schemas to postgres
	if err = store.PG.Migrator().CreateTable(
		&postgres.PaymentSchema{},
		&postgres.RequestSchema{},
		&postgres.WebhookRequestSchema{},
		&postgres.ShortCodeSchema{},
		&postgres.MpesaPaymentSchema{},
	); err != nil {
		return nil, err
	}

	return &Infrastructure{
		Conf:    conf,
		Storage: store,
		pg:      pg,
	}, nil

}

// ResetTables deletes all records in all tables. Can be used in conjunction
// with t.Cleanup() function after every test cases completes
func ResetTables(inf *Infrastructure) {
	l := logger.New(&logger.Config{LogMode: inf.Conf.LogLevel})
	l.Info().Msg("resetting tables")

	// clear tables of any data
	inf.Storage.PG.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&postgres.RequestSchema{})
	inf.Storage.PG.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&postgres.PaymentSchema{})
	inf.Storage.PG.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&postgres.WebhookRequestSchema{})
	inf.Storage.PG.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&postgres.ShortCodeSchema{})
	inf.Storage.PG.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&postgres.MpesaPaymentSchema{})

}

// CleanUp closes connection to database, then drops the test db.
// Only use after your test suite has fully completed
func CleanUp(inf *Infrastructure) {
	l := logger.New(&logger.Config{LogMode: inf.Conf.LogLevel})

	// close db
	inf.Storage.Close()

	sql := fmt.Sprintf(`DROP DATABASE "%s"`, inf.Conf.Postgres.DbName)
	l.Info().Str(logger.LData, sql).Msg("drop db sql")
	result := inf.pg.Exec(sql)
	if err := result.Error; err != nil {
		l.Error().Err(err).Msg("error dropping db")
		return
	}

}
