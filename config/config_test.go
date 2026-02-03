package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("QDRANT_URL", "http://localhost:6333")
	defer func() {
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("QDRANT_URL")
	}()

	os.Setenv("RPG_TELEGRAM_BOT_API_KEY", "test-bot-key")
	defer os.Unsetenv("RPG_TELEGRAM_BOT_API_KEY")

	tempConfig := `profile: "test"
inference_model:
  name: "gpt-4o-mini"
  url: "https://routerai.ru/v1"
  type: "routerai"
  api_key: "${OPENAI_API_KEY}"

embedding_model:
  name: "text-embedding-ada-002"
  url: "https://routerai.ru/v1"
  type: "routerai"
  api_key: "${OPENAI_API_KEY}"

vector_retriever:
  name: "qdrant"
  url: "${QDRANT_URL}"
  type: "qdrant"

telegram_bot_api_key: "${RPG_TELEGRAM_BOT_API_KEY}"`

	err := os.WriteFile("config.yaml", []byte(tempConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary config file: %v", err)
	}
	defer os.Remove("config.yaml")

	cfg, err := LoadConfig()
	if err != nil {
		t.Errorf("Failed to load config: %v", err)
		return
	}

	if cfg.Profile != "test" {
		t.Errorf("Expected profile 'test', got '%s'", cfg.Profile)
	}

	if cfg.InferenceModel.Name != "gpt-4o-mini" {
		t.Errorf("Expected inference model 'gpt-4o-mini', got '%s'", cfg.InferenceModel.Name)
	}

	if cfg.InferenceModel.Type != ModelTypeRouterAI {
		t.Errorf("Expected inference type RouterAI, got %v", cfg.InferenceModel.Type)
	}

	if cfg.EmbeddingModel.Name != "text-embedding-ada-002" {
		t.Errorf("Expected embedding model 'text-embedding-ada-002', got '%s'", cfg.EmbeddingModel.Name)
	}

	if cfg.VectorRetriever.Name != "qdrant" {
		t.Errorf("Expected retriever name 'qdrant', got '%s'", cfg.VectorRetriever.Name)
	}

	if cfg.TelegramBotApiKey != "${RPG_TELEGRAM_BOT_API_KEY}" {
		t.Errorf("Expected telegram bot api key from config, got '%s'", cfg.TelegramBotApiKey)
	}
}

func TestModelTypeString(t *testing.T) {
	tests := []struct {
		modelType ModelType
		expected  string
	}{
		{ModelTypeUnknown, "Unknown"},
		{ModelTypeOpenAI, "OpenAI"},
		{ModelTypeOllama, "Ollama"},
		{ModelTypeRouterAI, "RouterAI"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.modelType.String(); got != tt.expected {
				t.Errorf("ModelType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestModelTypeUnmarshalText(t *testing.T) {
	tests := []struct {
		input     string
		expected  ModelType
		wantError bool
	}{
		{"openai", ModelTypeOpenAI, false},
		{"ollama", ModelTypeOllama, false},
		{"routerai", ModelTypeRouterAI, false},
		{"invalid", ModelTypeUnknown, true},
		{"", ModelTypeUnknown, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var mt ModelType
			err := mt.UnmarshalText([]byte(tt.input))
			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("UnmarshalText() error = %v, wantError = false", err)
				return
			}
			if mt != tt.expected {
				t.Errorf("UnmarshalText() = %v, want %v", mt, tt.expected)
			}
		})
	}
}
