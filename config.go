package mongo

import (
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Config struct {
	MongoUser         string
	MongoPassword     string
	MongoUri          string
	Timeout           time.Duration
	databaseLimit     []string
	databaseOptions   *options.DatabaseOptions
	collectionOptions *options.CollectionOptions
	Debug             bool
}

func DefaultMongoConfig() *Config {
	return &Config{
		databaseOptions:   nil,
		databaseLimit:     []string{},
		collectionOptions: nil,
		Debug:             false,
	}
}
