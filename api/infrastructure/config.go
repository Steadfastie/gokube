package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/steadfastie/gokube/data/services"
	"go.uber.org/zap"
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
	env := os.Getenv("SETTINGS")
	if env == "" {
		logger.Error("SETTINGS is not set")
	}

	envFile := fmt.Sprintf("%sSettings.json", env)
	file, err := os.Open(envFile)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Error("Could not close file",
				zap.String("file", file.Name()),
				zap.Error(err))
		}
	}(file)

	config := &Config{}
	err = json.NewDecoder(file).Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}