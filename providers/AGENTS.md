# Providers Package

LLM backend implementations. Each provider implements `InferenceProvider` and/or `VectorEmbeddingProvider`.

## Structure

```
providers/
├── openai/              # OpenAI API (GPT models)
│   ├── openai.go        # OpenAIProvider, OpenAIEmbeddingProvider
│   └── openai_test.go
├── ollama/              # Local Ollama server
│   ├── ollama.go        # OllamaProvider, OllamaEmbeddingProvider
│   └── ollama_test.go
└── factory_integration_test.go  # Cross-provider tests
```

## Adding New Provider

```go
// 1. Create providers/{name}/{name}.go
package newprovider

import (
    "context"
    "go-llm-rpggamemaster/interfaces"
)

// 2. Compile-time interface check
var _ interfaces.InferenceProvider = (*NewProvider)(nil)

type NewProvider struct {
    model string
    // ... provider-specific fields
}

// 3. Constructor returns interface type
func NewNewProvider(model, apiKey, baseURL string) (interfaces.InferenceProvider, error) {
    // initialization...
    return &NewProvider{model: model}, nil
}

// 4. Implement interface methods
func (p *NewProvider) GenerateResponse(ctx context.Context, messages []interfaces.Message, temperature float64, maxTokens int) (string, error) {
    // implementation...
}

func (p *NewProvider) Name() string {
    return "newprovider"
}
```

## Interface Contract

```go
// InferenceProvider - LLM text generation
type InferenceProvider interface {
    GenerateResponse(ctx context.Context, messages []Message, temperature float64, maxTokens int) (string, error)
    Name() string
}

// VectorEmbeddingProvider - vector embeddings
type VectorEmbeddingProvider interface {
    EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)
    Name() string
}
```

## Provider Implementations

| Provider | Inference | Embedding | Default Model | Library |
|----------|-----------|-----------|---------------|---------|
| OpenAI | ✓ | ✓ | `gpt-3.5-turbo` / `text-embedding-ada-002` | langchaingo/llms/openai |
| Ollama | ✓ | ✓ | `llama3.1` / `nomic-embed-text` | langchaingo/llms/ollama |

## Integration Checklist

After creating provider:
1. Add `ModelType{Name}` to `config/model_type.go`
2. Add case to `factory/factory.go` switch statements
3. Add tests with interface compliance check
4. Update root `AGENTS.md` if significant

## Notes

- All providers use 60s timeout (hardcoded constant)
- `GenerateSimpleResponse()` exists but NOT in interface - helper method only
- Tests require running services (no mocks)
