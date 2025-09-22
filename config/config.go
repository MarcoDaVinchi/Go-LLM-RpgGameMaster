package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Profile           string          `mapstructure:"profile"`
	InferenceModel    LLModel         `mapstructure:"inference_model"`
	EmbeddingModel    LLModel         `mapstructure:"embedding_model"`
	VectorRetriever   VectorRetriever `mapstructure:"vector_retriever"`
	TelegramBotApiKey string          `mapstructure:"telegram_bot_api_key"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("config file not found: %v", err)
		} else {
			return nil, err
		}
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshall config: %w", err)
	}
	return &cfg, nil
}
