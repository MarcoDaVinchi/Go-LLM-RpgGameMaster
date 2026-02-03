package main

import (
	"context"
	"fmt"
	"log"

	"go-llm-rpggamemaster/config"
	"go-llm-rpggamemaster/factory"
	"go-llm-rpggamemaster/interfaces"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	providerFactory := factory.NewProviderFactory(cfg)

	llmProvider, err := providerFactory.CreateInferenceProvider()
	if err != nil {
		log.Fatalf("Failed to create LLM provider: %v", err)
	}

	fmt.Printf("Using LLM provider: %s\n", llmProvider.Name())

	embeddingProvider, err := providerFactory.CreateEmbeddingProvider()
	if err != nil {
		log.Fatalf("Failed to create embedding provider: %v", err)
	}

	fmt.Printf("Using embedding provider: %s\n", embeddingProvider.Name())

	ctx := context.Background()
	messages := []interfaces.Message{
		{Role: "user", Content: "Hello, how are you?"},
	}
	response, err := llmProvider.GenerateResponse(ctx, messages, 0.7, 0)
	if err != nil {
		log.Fatalf("Failed to generate response: %v", err)
	}

	fmt.Printf("LLM Response: %s\n", response)

	texts := []string{"Hello world", "How are you?"}
	embeddings, err := embeddingProvider.EmbedDocuments(ctx, texts)
	if err != nil {
		log.Fatalf("Failed to create embeddings: %v", err)
	}

	fmt.Printf("Created %d embeddings\n", len(embeddings))
}
