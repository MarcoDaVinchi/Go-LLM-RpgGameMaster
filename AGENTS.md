# Go LLM RPG Game Master

**Generated:** 2026-02-04 | **Commit:** 0508609 | **Branch:** main  
**Last Updated:** 2026-02-04

## ✅ Completed Migration

PostgreSQL + pgvector migration is **COMPLETE**. All 138 acceptance criteria met.

**Key Changes:**
- ✅ Migrated from Qdrant + SQLite to PostgreSQL + pgvector
- ✅ Implemented hybrid search (semantic + keyword with RRF fusion)
- ✅ ACID transactions between context and game state
- ✅ Dual-write strategy for zero-downtime migration
- ✅ Qdrant deprecated (kept for rollback safety)

## Overview

Telegram bot for RPG game mastering with LLM integration. Uses RouterAI.ru (OpenAI-compatible) via direct HTTP calls. PostgreSQL with pgvector extension for unified vector and game state storage.

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
│   ├── ollama/          # Ollama provider
│   └── routerai/        # RouterAI provider (primary)
├── retrievers/          # Vector storage (PostgreSQL + pgvector)
│   ├── postgres/        # PostgreSQL retriever (NEW - primary)
│   ├── dualwrite/       # Dual-write strategy (migration helper)
│   └── qdrant.go        # Qdrant retriever (DEPRECATED)
├── migrations/          # Database migrations
│   ├── 001_initial_schema.sql
│   └── 002_hybrid_search.sql
├── scripts/             # Migration scripts
│   ├── export-qdrant.sh
│   ├── convert-qdrant-to-postgres.go
│   ├── import-to-postgres.sh
│   └── validate-migration.sh
└── docs/                # Documentation
    ├── migration.md
    ├── benchmarks.md
    └── data-integrity-report.md
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
| Add retriever type | `config/retriever_type.go` | Follow `RetrieverType` pattern with iota |
| Database schema | `migrations/*.sql` | PostgreSQL schema and functions |
| Migration scripts | `scripts/` | Export, convert, import, validate |

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
| `RetrieverType` | type | `config/retriever_type.go:8` | Retriever enum (iota) |
| `PostgresRetriever` | struct | `retrievers/postgres/postgres.go:25` | PostgreSQL implementation |
| `DualWriteRetriever` | struct | `retrievers/dualwrite/dualwrite.go:18` | Dual-write strategy |

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
- String slicing in handlers (`Text[len("/cmd"):]`) lacks bounds checking
- `retrievers/qdrant.go` is deprecated but kept for rollback safety

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

# Database (PostgreSQL)
docker-compose up -d postgres   # Start PostgreSQL with pgvector
docker-compose down             # Stop PostgreSQL
psql $DATABASE_URL -f migrations/001_initial_schema.sql  # Apply schema
psql $DATABASE_URL -f migrations/002_hybrid_search.sql   # Apply functions
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
  url: "${DATABASE_URL:postgres://rpg:secret@localhost:5432/gamedb}"
  type: "postgres"  # or "qdrant" (deprecated)
```

### Required Environment Variables
```bash
RPG_TELEGRAM_BOT_API_KEY=...  # Telegram bot token
DATABASE_URL=postgres://rpg:secret@localhost:5432/gamedb
QDRANT_URL=http://localhost:6333  # Deprecated, kept for rollback
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

## Adding New Retriever

1. Create `retrievers/{name}/{name}.go` or `retrievers/{name}.go`
2. Implement `Retriever` interface
3. Add compile-time check: `var _ interfaces.Retriever = (*Provider)(nil)`
4. Add `RetrieverType{Name}` to `config/retriever_type.go` with `iota`
5. Add case in `factory/factory.go` CreateRetriever switch statement
6. Add tests in `retrievers/{name}/{name}_test.go`

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
- PostgreSQL with pgvector: `pgvector/pgvector:pg16` image
- Tests require running services (PostgreSQL, Ollama) - no mocks implemented
- Go 1.25.6 required (see go.mod)
- Migration complete: Qdrant deprecated, PostgreSQL is primary
- Dual-write available for zero-downtime transitions
