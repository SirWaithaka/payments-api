package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/SirWaithaka/payments-api/internal/config"
)

func NewLogger() *Logger {
	return &Logger{
		SlowThreshold:             300 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
	}
}

type Logger struct {
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

func (l *Logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	logger := zerolog.Ctx(ctx)
	logger.Info().Interface("data", data).Msg(msg)
}

func (l *Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	logger := zerolog.Ctx(ctx)
	logger.Warn().Interface("data", data).Msg(msg)
}

func (l *Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	logger := zerolog.Ctx(ctx)
	logger.Error().Interface("data", data).Msg(msg)
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	logger := zerolog.Ctx(ctx)

	elapsed := time.Since(begin)
	sql, rows := fc()
	lg := logger.With().Str("sql_duration", elapsed.String()).Logger()
	var lo *zerolog.Event
	switch {
	case err != nil && (!errors.Is(err, gorm.ErrRecordNotFound)):
		lo = lg.Error()
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		lo = lg.Warn().Str("slowSql", slowLog)
	default:
		lo = lg.Info()
	}
	if rows == -1 {
		lo.Msg(sql)
		return
	}
	lo.Int64("rows", rows).Msg(sql)
}

// NewPostgresClient creates a new connection to postgres
func NewPostgresClient(cfg config.PostgresConfigs) (*gorm.DB, error) {
	var err error

	gormconfig := gorm.Config{Logger: NewLogger(), TranslateError: false}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{DSN: postgresDSN(cfg)}), &gormconfig)
	if err != nil {
		return nil, fmt.Errorf("gorm open error: %v", err)
	}

	// get native database/sql connection
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("native sql fetch error: %v", err)
	}

	// test connection to db
	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping error: %v", err)
	}

	return gormDB, nil
}

func postgresDSN(config config.PostgresConfigs) string {
	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode='disable' search_path=%s",
		config.User, config.Password, config.DbName, config.Host, config.Port, config.Schema,
	)
}
