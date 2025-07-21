package storage

import (
	"gorm.io/gorm"

	"github.com/SirWaithaka/payments-api/internal/config"
)

// Database is a wrapper type for the gorm DB pointer
type Database struct {
	PG *gorm.DB
	//Cache *redis.Redis
}

func (db *Database) Close() {
	// close connection to postgres db
	sqlI, _ := db.PG.DB()
	if sqlI != nil {
		_ = sqlI.Close()
	}

	//// close redis client
	//db.Cache.Client().Close()
}

// NewDatabase return a Database instance
func NewDatabase(config config.Config) (*Database, error) {
	// connect and return postgres client
	postgresClient, err := NewPostgresClient(config.Postgres)
	if err != nil {
		return nil, err
	}

	// connect and create redis connection
	//redisClient := redis.New(config.Redis.BaseURL)
	// capture any panics and log

	return &Database{PG: postgresClient}, nil
}
