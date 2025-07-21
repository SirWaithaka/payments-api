package config

import "github.com/kelseyhightower/envconfig"

type envConfig struct {
	LogLevel string `envconfig:"log_level" default:"debug"`
	HTTPPort string `envconfig:"http_port" default:"6000"`

	PostgresUser     string `envconfig:"postgres_user" required:"true"`
	PostgresPassword string `envconfig:"postgres_password" required:"true"`
	PostgresHost     string `envconfig:"postgres_host" required:"true"`
	PostgresPort     string `envconfig:"postgres_port" required:"true"`
	PostgresDBName   string `envconfig:"postgres_database" default:"payments"`
	PostgresSchema   string `envconfig:"postgres_schema" default:"public"`
}

func FromEnv(cfg *Config) error {
	var c envConfig
	if err := envconfig.Process("", &c); err != nil {
		return err
	}

	cfg.LogLevel = c.LogLevel
	cfg.HTTPPort = c.HTTPPort

	cfg.Postgres.User = c.PostgresUser
	cfg.Postgres.Password = c.PostgresPassword
	cfg.Postgres.Host = c.PostgresHost
	cfg.Postgres.Port = c.PostgresPort
	cfg.Postgres.DbName = c.PostgresDBName
	cfg.Postgres.Schema = c.PostgresSchema

	return nil
}
