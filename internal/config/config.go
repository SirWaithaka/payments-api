package config

type PostgresConfigs struct {
	User     string
	Password string
	Host     string
	Port     string
	DbName   string
	Schema   string
}

type Config struct {
	ServiceName string
	LogLevel    string
	HTTPPort    string
	Postgres    PostgresConfigs
}
