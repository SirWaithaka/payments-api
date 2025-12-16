package config

import (
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/pkg/logger"
)

type PostgresConfigs struct {
	User     string
	Password string
	Host     string
	Port     string
	DbName   string
	Schema   string
}

type KafkaConfig struct {
	Host string
}

type DarajaConfig struct {
	Endpoint string
}

type QuikkConfig struct {
	Endpoint string
}

type Config struct {
	logger *zerolog.Logger

	ServiceName string
	LogLevel    string
	HTTPPort    string
	Postgres    PostgresConfigs
	Kafka       KafkaConfig
	Daraja      DarajaConfig
	Quikk       QuikkConfig
}

func (c Config) Logger() *zerolog.Logger {
	if c.logger == nil {
		l := logger.New(&logger.Config{LogMode: c.LogLevel, Service: c.ServiceName})
		c.logger = &l
		return c.logger
	}
	return c.logger
}
