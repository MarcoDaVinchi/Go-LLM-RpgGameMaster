package routerai

import (
	"testing"

	"go-llm-rpggamemaster/interfaces"
)

func TestNewRouterAIProvider(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		apiKey    string
		baseURL   string
		wantError bool
	}{
		{
			name:      "valid provider with model",
			model:     "gpt-4o-mini",
			apiKey:    "test-key",
			baseURL:   "",
			wantError: false,
		},
		{
			name:      "missing model returns error",
			model:     "",
			apiKey:    "test-key",
			baseURL:   "",
			wantError: true,
		},
		{
			name:      "custom base URL",
			model:     "gpt-4o-mini",
			apiKey:    "test-key",
			baseURL:   "https://custom.router.ai/v1",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewRouterAIProvider(tt.model, tt.apiKey, tt.baseURL)
			if tt.wantError {
				if err == nil {
					t.Errorf("NewRouterAIProvider() error = nil, wantError = true")
				}
				return
			}
			if err != nil {
				t.Errorf("NewRouterAIProvider() error = %v, wantError = false", err)
				return
			}
			if provider == nil {
				t.Error("NewRouterAIProvider() returned nil provider")
			}
		})
	}
}

func TestRouterAIProvider_Name(t *testing.T) {
	provider, err := NewRouterAIProvider("gpt-4o-mini", "test-key", "")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if name := provider.Name(); name != "routerai" {
		t.Errorf("Name() = %v, want %v", name, "routerai")
	}
}

func TestRouterAIProvider_InterfaceCompliance(t *testing.T) {
	var (
		_ interfaces.InferenceProvider       = (*RouterAIProvider)(nil)
		_ interfaces.VectorEmbeddingProvider = (*RouterAIProvider)(nil)
	)
}
