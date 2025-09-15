package openai

import (
	"testing"

	"go-llm-rpggamemaster/interfaces"
)

func TestOpenAIProvider(t *testing.T) {
	// Check that the provider implements the LLMProvider interface
	var _ interfaces.LLMProvider = &OpenAIProvider{}

	// Check that the embedding provider implements the EmbeddingProvider interface
	var _ interfaces.EmbeddingProvider = &OpenAIEmbeddingProvider{}
}
