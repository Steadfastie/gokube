package infrastructure

import (
	"os"
	"strings"

	"github.com/steadfastie/gokube/data/errors"
	"github.com/steadfastie/gokube/data/services"
)

const (
	EnvMongoConnectionString = "MONGO_CONNECTION_STRING"
	EnvMongoDatabase         = "MONGO_DATABASE"
	EnvLogLevel              = "LOGLEVEL"
	EnvCron                  = "CRON"
	EnvKafkaAddresses        = "KAFKA_ADDRESSES"
)

type Config struct {
	MongoSettings services.MongoSettings
	LogLevel      string
	Cron          string
	KafkaServers  []string
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
		logLevel = "Information"
	}

	cronExpression := os.Getenv(EnvCron)
	if cronExpression == "" {
		cronExpression = "*/5 * * * * *" // Defaults to every 5 seconds
	}

	kafkaBootstrapServer := os.Getenv(EnvKafkaAddresses)
	addresses := []string{}
	if kafkaBootstrapServer == "" {
		addresses = append(addresses, "localhost")
	} else {
		addresses = append(addresses, strings.Split(kafkaBootstrapServer, ",")...)
	}

	config := &Config{
		MongoSettings: services.MongoSettings{
			ConnectionString: mongoConnectionString,
			Database:         mongoDatabase,
		},
		LogLevel:     logLevel,
		Cron:         cronExpression,
		KafkaServers: addresses,
	}

	return config, nil
}
