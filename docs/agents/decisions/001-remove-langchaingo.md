# ADR 001: Remove langchaingo Dependency

## Status
**Accepted** (2026-02-03)

**Decision Owner:** @Sisyphus (AI Agent)  
**Stakeholders:** Project maintainers

---

## Context

The project currently uses `github.com/tmc/langchaingo` as an abstraction layer for LLM interactions. Over time, several issues have emerged:

1. **Weak maintainability:** langchaingo has infrequent updates and limited community support
2. **Hidden complexity:** Simple HTTP requests to LLM APIs are wrapped in heavy abstraction
3. **Overhead:** The library adds ~53+ indirect dependencies without providing significant value
4. **Limited control:** Abstracted interfaces make it harder to customize API calls or handle edge cases

The project's actual needs are simple:
- HTTP POST to `/v1/chat/completions` for text generation
- HTTP POST to `/v1/embeddings` for vector embeddings
- Basic authentication (API key header)

These are straightforward operations that don't require a heavy abstraction library.

---

## Decision

**Remove langchaingo and implement direct HTTP calls using standard library (`net/http`).**

### Rationale

1. **Simplicity:** Direct HTTP calls are more transparent and easier to debug
2. **Control:** Full access to request/response handling
3. **Dependencies:** Reduces dependency tree by ~53 packages
4. **Maintainability:** Less abstraction means less magic, easier to understand
5. **Standard patterns:** HTTP clients are well-understood by all Go developers

---

## Consequences

### Positive ➕

- **Reduced dependencies:** ~53 fewer indirect dependencies in go.mod
- **Full control:** Direct access to HTTP requests, headers, timeouts
- **Better observability:** Easier to add logging, metrics, tracing
- **Simpler debugging:** No hidden layers between code and API
- **Standard patterns:** Uses `net/http` which all Go developers know
- **Smaller binary:** Less vendored code

### Negative ➖

- **More boilerplate:** ~100 lines of HTTP client code per provider
- **Manual error handling:** Need to parse HTTP errors and JSON responses
- **Type definitions:** Must define own structs for requests/responses
- **Retry logic:** Need to implement manually (if needed)

### Neutral ➡️

- **Testing:** Same effort (HTTP calls need mocking either way)
- **Configuration:** Same complexity

---

## Implementation Notes

### Affected Components

| Component | Current | New |
|-----------|---------|-----|
| `providers/openai/` | Uses `langchaingo/llms/openai` | Remove, use `providers/routerai/` |
| `providers/ollama/` | Uses `langchaingo/llms/ollama` | Remove, use `providers/routerai/` |
| `retrievers/qdrant.go` | Uses `langchaingo/vectorstores/qdrant` | Use Qdrant HTTP API or official Go client |
| `retrievers/sqlite.go` | Uses `langchaingo/schema` | Define own `Document` type |
| `factory/factory.go` | Uses `langchaingo/embeddings` | Update to use new types |

### Implementation Steps

1. Create `providers/routerai/` with HTTP client
2. Implement `InferenceProvider` interface
3. Implement `VectorEmbeddingProvider` interface
4. Update factory to instantiate RouterAI provider
5. Refactor retrievers to remove langchaingo types
6. Run `go mod tidy` to remove unused dependencies
7. Update and run tests

### Code Example

**Before (with langchaingo):**
```go
import "github.com/tmc/langchaingo/llms/openai"

llm, err := openai.New(
    openai.WithToken(apiKey),
    openai.WithModel("gpt-4"),
)
response, err := llm.GenerateContent(ctx, messages)
```

**After (direct HTTP):**
```go
type RouterAIProvider struct {
    client  *http.Client
    apiKey  string
    model   string
    baseURL string
}

func (p *RouterAIProvider) GenerateResponse(ctx context.Context, messages []Message, temperature float64, maxTokens int) (string, error) {
    req := ChatCompletionRequest{
        Model:    p.model,
        Messages: convertMessages(messages),
        Temperature: temperature,
        MaxTokens: maxTokens,
    }
    // Standard HTTP POST
    resp, err := p.doRequest(ctx, "/v1/chat/completions", req)
    // ...
}
```

---

## Related Decisions

- [ADR 002: RouterAI Provider](./002-routerai-provider.md) - Choice of OpenAI-compatible provider

---

## References

- langchaingo repository: https://github.com/tmc/langchaingo
- OpenAI API Reference: https://platform.openai.com/docs/api-reference
- Go net/http documentation: https://pkg.go.dev/net/http
