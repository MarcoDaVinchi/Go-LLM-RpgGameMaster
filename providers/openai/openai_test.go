package openai

import (
	"testing"

	"go-llm-rpggamemaster/interfaces"
)

func TestOpenAIProvider(t *testing.T) {
	// Check that the provider implements the InferenceProvider interface
	var _ interfaces.InferenceProvider = &OpenAIProvider{}

	// Check that the embedding provider implements the VectorEmbeddingProvider interface
	var _ interfaces.VectorEmbeddingProvider = &OpenAIEmbeddingProvider{}
}
