package config

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
	ServiceName string
	LogLevel    string
	HTTPPort    string
	Postgres    PostgresConfigs
	Kafka       KafkaConfig
	Daraja      DarajaConfig
	Quikk       QuikkConfig
}
