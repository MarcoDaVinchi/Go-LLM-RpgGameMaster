# Go-LLM-RPGGameMaster Provider Model

This project implements a generic provider model for LLM and embedding services, allowing for easy integration of multiple providers like Ollama and OpenAI. The legacy client implementations in `clients/` and `llm/` are deprecated and will be removed in a future version.

## Package Structure

```
├── config/                 # Configuration structures and parsing
├── factory/                # Provider factory implementation
│   └── interface/          # Factory interface definitions
├── providers/              # Generic provider interfaces
│   ├── ollama/             # Ollama provider implementation
│   └── openai/             # OpenAI provider implementation
└── examples/               # Example usage
```

## Configuration

The provider model uses a YAML-based configuration system. Create a `config.yaml` file in the root directory:

```yaml
llm_providers:
  - name: "ollama"
    model: "llama3.1"
  - name: "openai"
    model: "gpt-3.5-turbo"
    api_key: "${OPENAI_API_KEY}"

embedding_providers:
  - name: "ollama"
    model: "nomic-embed-text"
  - name: "openai"
    model: "text-embedding-ada-002"
    api_key: "${OPENAI_API_KEY}"

default_llm: "ollama"
default_embedding: "ollama"
```

## Usage

### Creating Providers

```go
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

// Create embedding provider
embeddingProvider, err := providerFactory.CreateEmbeddingProvider("ollama", "nomic-embed-text")
if err != nil {
    log.Fatalf("Failed to create embedding provider: %v", err)
}
```

### Using LLM Providers

```go
// Generate a simple response
ctx := context.Background()
response, err := llmProvider.GenerateSimpleResponse(ctx, "Hello, how are you?")
if err != nil {
    log.Fatalf("Failed to generate response: %v", err)
}

fmt.Printf("LLM Response: %s\n", response)
```

### Using Embedding Providers

```go
// Create embeddings
texts := []string{"Hello world", "How are you?"}
embeddings, err := embeddingProvider.EmbedDocuments(ctx, texts)
if err != nil {
    log.Fatalf("Failed to create embeddings: %v", err)
}

fmt.Printf("Created %d embeddings\n", len(embeddings))
```

## Provider Interface

### LLMProvider

```go
type LLMProvider interface {
    GenerateResponse(ctx context.Context, messages []Message, temperature float64, maxTokens int) (string, error)
    GenerateSimpleResponse(ctx context.Context, userMessage string) (string, error)
    Name() string
}
```

### EmbeddingProvider

```go
type EmbeddingProvider interface {
    EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)
    Name() string
}
```

## Supported Providers

1. **Ollama** - For local LLM and embedding models
2. **OpenAI** - For OpenAI API models

## Extending with New Providers

To add a new provider:

1. Create a new subdirectory in `providers/` with the provider name
2. Implement the `LLMProvider` and/or `EmbeddingProvider` interfaces
3. Update the factory to support creating instances of your new provider
4. Add the provider configuration to `config.yaml`

## Backward Compatibility

The new provider model is designed to be backward compatible with existing functionality. The legacy client implementations in `clients/` and `llm/` are deprecated but can still be used alongside the new provider model. These deprecated implementations will be removed in a future version. We recommend migrating to the new provider model as soon as possible.
