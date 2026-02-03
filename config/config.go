package config

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Config struct {
	Profile           string          `mapstructure:"profile"`
	InferenceModel    LLModel         `mapstructure:"inference_model"`
	EmbeddingModel    LLModel         `mapstructure:"embedding_model"`
	VectorRetriever   VectorRetriever `mapstructure:"vector_retriever"`
	TelegramBotApiKey string          `mapstructure:"telegram_bot_api_key"`
}

func decodeHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}

	switch t {
	case reflect.TypeOf(ModelType(0)):
		var mt ModelType
		if err := mt.UnmarshalText([]byte(data.(string))); err != nil {
			return nil, err
		}
		return mt, nil
	case reflect.TypeOf(RetrieverType(0)):
		var rt RetrieverType
		if err := rt.UnmarshalText([]byte(data.(string))); err != nil {
			return nil, err
		}
		return rt, nil
	}

	return data, nil
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
		}
		return nil, err
	}

	var cfg Config
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: decodeHook,
		Result:     &cfg,
	}

	if err := viper.Unmarshal(&cfg, viper.DecodeHook(decodeHook)); err != nil {
		return nil, fmt.Errorf("unmarshall config: %w", err)
	}

	_ = decoderConfig
	return &cfg, nil
}
