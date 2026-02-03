# Go LLM RPG Game Master

**Generated:** 2026-02-03 | **Commit:** 71f8f47 | **Branch:** main  
**Last Updated:** 2026-02-03

## ⚠️ Active Migration Context

🚨 **IMPORTANT:** Project is currently migrating away from langchaingo to direct HTTP calls.  
📋 **See:** [AGENTS-TODO.md](./AGENTS-TODO.md) for current status and next steps.

**Key Changes:**
- Removing `langchaingo` dependency (weak maintenance, hides simple HTTP)
- Consolidating to single provider: **RouterAI.ru** (OpenAI-compatible API)
- OpenAI and Ollama providers will be replaced by unified `providers/routerai/`

## Overview

Telegram bot for RPG game mastering with LLM integration. Uses RouterAI.ru (OpenAI-compatible) via direct HTTP calls. Qdrant/SQLite for vector storage.

## Structure

```
./
├── main.go              # Entry point, Telegram bot handlers
├── config/              # Config loading (viper/YAML), custom enum types
├── factory/             # Provider factory pattern
│   └── interface/       # Factory interface definition
├── interfaces/          # Core interfaces (InferenceProvider, VectorEmbeddingProvider)
├── providers/           # LLM backend implementations
│   ├── openai/          # OpenAI provider
│   └── ollama/          # Ollama provider
├── retrievers/          # Vector storage (Qdrant, SQLite)
└── qdrant_storage/      # Qdrant data (not Go code)
```

## Where to Look

| Task | Location | Notes |
|------|----------|-------|
| Add Telegram command | `main.go:52-54` | Register handler with `b.RegisterHandler()` |
| Add LLM provider | `providers/{name}/` | Implement `InferenceProvider` interface |
| Change config | `config/config.go` | Add field to `Config` struct, update YAML |
| Add enum type | `config/model_type.go` | Follow `ModelType` pattern with iota |
| Add retriever | `retrievers/` | Implement `Retriever` interface |
| Factory changes | `factory/factory.go` | Add case to switch statement |

## Code Map

| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `InferenceProvider` | interface | `interfaces/providers.go:8` | LLM generation contract |
| `VectorEmbeddingProvider` | interface | `interfaces/providers.go:16` | Embedding contract |
| `ProviderFactory` | interface | `factory/interface/factory.go:9` | Factory contract |
| `NewProviderFactory` | func | `factory/factory.go:21` | Factory constructor |
| `OpenAIProvider` | struct | `providers/openai/openai.go:19` | OpenAI implementation |
| `OllamaProvider` | struct | `providers/ollama/ollama.go:19` | Ollama implementation |
| `LoadConfig` | func | `config/config.go:18` | YAML config loader |
| `ModelType` | type | `config/model_type.go:8` | Provider enum (iota) |

## Conventions

### Imports
```go
import (
    "context"           // stdlib first
    "fmt"
    
    "github.com/rs/zerolog/log"  // third-party second
    
    "go-llm-rpggamemaster/interfaces"  // internal last
)
```

### Naming
- Packages: `lowercase` single word
- Interfaces: `-er` suffix (`InferenceProvider`)
- Structs: `PascalCase` (`OpenAIProvider`)
- Exported: `PascalCase`, unexported: `camelCase`
- Enums: `iota` with `String()`, `MarshalText()`, `UnmarshalText()`

### Error Handling
```go
if err != nil {
    return fmt.Errorf("context: %w", err)  // wrap with context
}
log.Err(err).Msg("description")            // zerolog for logging
log.Fatal().Err(err).Msg("...")            // Fatal only in main()
```

### Provider Pattern
```go
// 1. Interface definition
type InferenceProvider interface {
    GenerateResponse(ctx context.Context, messages []Message, temperature float64, maxTokens int) (string, error)
    Name() string
}

// 2. Compile-time check
var _ interfaces.InferenceProvider = (*OpenAIProvider)(nil)

// 3. Constructor returns interface
func NewOpenAIProvider(...) (interfaces.InferenceProvider, error)
```

### Testing
- Same package (`package openai` not `package openai_test`)
- Table-driven with `t.Run()`
- Interface compliance: `var _ Interface = &Struct{}`
- No external frameworks (stdlib `testing` only)

## Anti-Patterns

### NEVER
- Use `clients/` or `llm/` packages - **DEPRECATED**, use `providers/`
- Call external APIs in tests without mocks
- Use `as any` or type suppression
- Leave empty catch blocks
- `panic()` outside main - use `log.Fatal()`

### Hardcoded Values (Technical Debt)
- `60 * time.Second` timeout in providers - should be configurable
- Default models (`llama3.1`, `gpt-3.5-turbo`) - consider config
- Temperature `0.7` in `GenerateSimpleResponse` - should be parameter

### Known Issues
- `factory.CreateRetriever()` has undefined variables (`embedder`, `retrieverType`) - incomplete
- String slicing in handlers (`Text[len("/cmd"):]`) lacks bounds checking

## Commands

```bash
# Build & Run
go build -o bin/app .
go run main.go

# Test
go test ./...                              # all
go test ./providers/openai -run TestName   # specific
go test -v -cover ./...                    # verbose + coverage

# Lint
go fmt ./...
go vet ./...
golangci-lint run                          # if installed

# Database (via task)
task create-db    # Create with schema
task seed-db      # Add test data
task reset-db     # Clean + create + seed
task shell        # SQLite shell
```

## Configuration

### config.yml (gitignored)
```yaml
profile: "local"  # "local" = pretty logs, "prod" = JSON

inference_model:
  url: "${INFERENCE_SERVER_URL:https://api.openai.com/v1}"
  type: "openai"  # or "ollama"
  name: "gpt-4o-mini"
  api_key: "${OPENAI_API_KEY}"

embedding_model:
  url: "${EMBEDDING_SERVER_URL:http://localhost:11434}"
  type: "ollama"
  name: "nomic-embed-text"

vector_retriever:
  url: "${QDRANT_URL:http://localhost:6333}"
  type: "qdrant"
```

### Required Environment Variables
```bash
RPG_TELEGRAM_BOT_API_KEY=...  # Telegram bot token
QDRANT_URL=http://localhost:6333
OPENAI_API_KEY=...            # If using OpenAI
INFERENCE_SERVER_URL=...      # Optional, has defaults
EMBEDDING_SERVER_URL=...      # Optional, has defaults
```

## Adding New Provider

1. Create `providers/{name}/{name}.go`
2. Implement `InferenceProvider` and/or `VectorEmbeddingProvider`
3. Add compile-time check: `var _ interfaces.InferenceProvider = (*Provider)(nil)`
4. Add `ModelType{Name}` to `config/model_type.go` with `iota`
5. Add case in `factory/factory.go` switch statements
6. Add tests in `providers/{name}/{name}_test.go`

## Documentation for AI Agents

This project uses structured documentation for AI-assisted development:

| File | Purpose |
|------|---------|
| `AGENTS.md` | This file - general project overview and conventions |
| `AGENTS-TODO.md` | Active TODOs, migration plans, current context |
| `docs/agents/decisions/` | Architecture Decision Records (ADRs) |
| `docs/agents/context/` | Known issues, technical debt, context notes |

### Key ADRs

- [ADR 001: Remove langchaingo](./docs/agents/decisions/001-remove-langchaingo.md)
- [ADR 002: RouterAI Provider](./docs/agents/decisions/002-routerai-provider.md)

### For Future Agents

**Starting a new session?** Read these in order:
1. Check `AGENTS-TODO.md` for current active work
2. Review relevant ADRs in `docs/agents/decisions/`
3. Check `docs/agents/context/known-issues.md` for pitfalls
4. Then proceed with implementation

## Notes

- Ollama timeout: 60s hardcoded (may need increase for large models)
- Default embedding: `nomic-embed-text` for Ollama, `text-embedding-ada-002` for OpenAI
- Qdrant collection: `game_collection` (hardcoded in `retrievers/qdrant.go`)
- Tests require running services (Ollama, Qdrant) - no mocks implemented
- Go 1.25.6 required (see go.mod)
