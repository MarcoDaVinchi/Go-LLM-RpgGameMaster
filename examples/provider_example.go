package main

import (
	"context"
	"fmt"
	"log"

	"go-llm-rpggamemaster/config"
	factory "go-llm-rpggamemaster/factory"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfigFromDefaultPath()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create provider factory
	providerFactory := factory.NewProviderFactory(cfg)

	// Create LLM provider
	llmProvider, err := providerFactory.CreateLLMProvider("ollama", "llama3.1")
	if err != nil {
		log.Fatalf("Failed to create LLM provider: %v", err)
	}

	fmt.Printf("Using LLM provider: %s\n", llmProvider.Name())

	// Create embedding provider
	embeddingProvider, err := providerFactory.CreateEmbeddingProvider("ollama", "nomic-embed-text")
	if err != nil {
		log.Fatalf("Failed to create embedding provider: %v", err)
	}

	fmt.Printf("Using embedding provider: %s\n", embeddingProvider.Name())

	// Example usage of LLM provider
	ctx := context.Background()
	response, err := llmProvider.GenerateSimpleResponse(ctx, "Hello, how are you?")
	if err != nil {
		log.Fatalf("Failed to generate response: %v", err)
	}

	fmt.Printf("LLM Response: %s\n", response)

	// Example usage of embedding provider
	texts := []string{"Hello world", "How are you?"}
	embeddings, err := embeddingProvider.EmbedDocuments(ctx, texts)
	if err != nil {
		log.Fatalf("Failed to create embeddings: %v", err)
	}

	fmt.Printf("Created %d embeddings\n", len(embeddings))

	// Example of creating OpenAI provider
	openaiProvider, err := providerFactory.CreateLLMProvider("openai", "gpt-3.5-turbo")
	if err != nil {
		log.Printf("Failed to create OpenAI provider: %v", err)
	} else {
		fmt.Printf("Successfully created OpenAI provider: %s\n", openaiProvider.Name())
	}
}
