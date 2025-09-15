package ollama

import (
	"context"
	"testing"

	"go-llm-rpggamemaster/interfaces"
)

func TestOllamaProvider(t *testing.T) {
	// Create a new Ollama provider
	provider, err := NewOllamaProvider("llama3.1", "")
	if err != nil {
		t.Fatalf("Failed to create Ollama provider: %v", err)
	}

	// Check that the provider implements the LLMProvider interface
	if _, ok := provider.(interfaces.LLMProvider); !ok {
		t.Error("OllamaProvider does not implement LLMProvider interface")
	}

	// Check that the provider has the correct name
	if provider.Name() != "ollama" {
		t.Errorf("Expected provider name 'ollama', got '%s'", provider.Name())
	}

	// Test simple response generation (this would require a running Ollama instance)
	// For testing purposes, we'll just check that the method exists
	ctx := context.Background()
	_, err = provider.GenerateSimpleResponse(ctx, "Hello, world!")
	// We don't check the error here because it would require a running Ollama instance
	// The fact that it compiles and runs without panicking is sufficient for this test
	_ = err
}
