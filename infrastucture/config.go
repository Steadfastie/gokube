package infrastucture

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"os"
)

type Config struct {
	MongoSettings MongoSettings `json:"MongoSettings"`
	LogLevel      string        `json:"LogLevel"`
}

type MongoSettings struct {
	MongoConnectionString string `json:"MongoConnectionString"`
	MongoDatabase         string `json:"MongoDatabase"`
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
