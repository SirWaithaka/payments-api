package config

import "github.com/kelseyhightower/envconfig"

type envConfig struct {
	LogLevel string `envconfig:"log_level" default:"debug"`
	HTTPPort string `envconfig:"http_port" default:"6000"`
}

func FromEnv(cfg *Config) error {
	var c envConfig
	if err := envconfig.Process("", &c); err != nil {
		return err
	}

	cfg.LogLevel = c.LogLevel
	cfg.HTTPPort = c.HTTPPort

	return nil
}
