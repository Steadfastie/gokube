package infrastructure

import (
	"context"
	"os"

	"github.com/steadfastie/gokube/data/errors"
	"github.com/steadfastie/gokube/data/services"
	"go.uber.org/zap"
)

const (
	EnvAuthDomain            = "AUTH_DOMAIN"
	EnvAuthAudience          = "AUTH_AUDIENCE"
	EnvMongoConnectionString = "MONGO_CONNECTION_STRING"
	EnvMongoDatabase         = "MONGO_DATABASE"
	EnvLogLevel              = "LOGLEVEL"
)

type Config struct {
	Auth          AuthSettings           `json:"Auth0"`
	MongoSettings services.MongoSettings `json:"MongoSettings"`
	LogLevel      string                 `json:"LogLevel"`
}

func (c *Config) GetMongoSettings() services.MongoSettings {
	return c.MongoSettings
}

func (c *Config) GetLogLevel() string {
	return c.LogLevel
}

type AuthSettings struct {
	Domain   string `json:"Domain"`
	Audience string `json:"Audience"`
}

func NewConfig(ctx context.Context, logger *zap.Logger) (*Config, error) {
	authDomain := os.Getenv(EnvAuthDomain)
	if authDomain == "" {
		panic(errors.NewBusinessRuleError("Auth domain is not set"))
	}

	authAudience := os.Getenv(EnvAuthAudience)
	if authAudience == "" {
		panic(errors.NewBusinessRuleError("Auth audience is not set"))
	}

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

	config := &Config{
		Auth: AuthSettings{
			Domain:   authDomain,
			Audience: authAudience,
		},
		MongoSettings: services.MongoSettings{
			ConnectionString: mongoConnectionString,
			Database:         mongoDatabase,
		},
		LogLevel: logLevel,
	}

	return config, nil
}
