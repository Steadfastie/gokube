package infrastructure

import (
	"os"

	"github.com/steadfastie/gokube/data/errors"
	"github.com/steadfastie/gokube/data/services"
)

const (
	EnvMongoConnectionString = "MONGO_CONNECTION_STRING"
	EnvMongoDatabase         = "MONGO_DATABASE"
	EnvLogLevel              = "LOGLEVEL"
	EnvCron                  = "CRON"
)

type Config struct {
	MongoSettings services.MongoSettings
	LogLevel      string
	Cron          string
}

func (config *Config) GetMongoSettings() services.MongoSettings {
	return config.MongoSettings
}

func (config *Config) GetLogLevel() string {
	return config.LogLevel
}

func NewConfig() (*Config, error) {
	mongoConnectionString := os.Getenv(EnvMongoConnectionString)
	if mongoConnectionString == "" {
		panic(errors.NewBusinessRuleError("Mongo connection string is not set"))
	}

	mongoDatabase := os.Getenv(EnvMongoDatabase)
	if mongoDatabase == "" {
		panic(errors.NewBusinessRuleError("Mongo database is not set"))
	}

	logLevel := os.Getenv(EnvLogLevel)
	if logLevel == "" {
		logLevel = "Information" // Defaults to Information
	}

	cronExpression := os.Getenv(EnvCron)
	if cronExpression == "" {
		cronExpression = "*/5 * * * * *" // Defaults to every 5 seconds
	}

	config := &Config{
		MongoSettings: services.MongoSettings{
			ConnectionString: mongoConnectionString,
			Database:         mongoDatabase,
		},
		LogLevel: logLevel,
		Cron:     cronExpression,
	}

	return config, nil
}
