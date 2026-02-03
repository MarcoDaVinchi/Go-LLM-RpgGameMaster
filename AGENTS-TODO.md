# Active TODOs & Migration Plans

**Last Updated:** 2026-02-03  
**Current Focus:** Remove langchaingo dependency, migrate to RouterAI.ru

---

## 🚀 [MIGRATION-001] Remove langchaingo → Direct HTTP

**Status:** ✅ COMPLETED  
**Started:** 2026-02-03  
**Completed:** 2026-02-03  
**Priority:** High

### Context
Migration from langchaingo library to direct HTTP calls. Langchaingo hides simple HTTP interactions behind heavy abstraction layer with weak maintainability. RouterAI.ru (OpenAI-compatible API) will be the single LLM provider.

### Completed ✅
- [x] Assessed langchaingo usage across codebase
- [x] Evaluated replacement complexity (LOW)
- [x] Selected RouterAI.ru as unified provider (OpenAI-compatible)
- [x] Created migration documentation
- [x] Create `providers/routerai/` package
- [x] Implement HTTP client for RouterAI.ru
- [x] Add chat completion endpoint (`POST /v1/chat/completions`)
- [x] Add embeddings endpoint (`POST /v1/embeddings`)
- [x] Update factory to use RouterAI provider
- [x] Remove `providers/openai/` (replaced by routerai)
- [x] Remove `providers/ollama/` (replaced by routerai)
- [x] Refactor retrievers to remove langchaingo dependencies
- [x] Update `go.mod` - remove langchaingo
- [x] Run `go mod tidy`
- [x] Update tests
- [x] Verify all tests pass
- [x] Update AGENTS.md with new architecture

### Summary
**Migration completed successfully!**

**Key Changes:**
1. **New Provider:** `providers/routerai/` - Direct HTTP implementation for RouterAI.ru
2. **Removed Dependencies:** langchaingo + ~53 indirect dependencies removed from go.mod
3. **Unified Provider:** Single RouterAI.ru provider instead of OpenAI + Ollama
4. **Refactored Retrievers:** Qdrant and SQLite retrievers now use direct HTTP instead of langchaingo types
5. **Updated Factory:** Now creates RouterAI provider, old providers return deprecation errors

**Verification:**
- ✅ `go build ./...` - Success
- ✅ `go test ./...` - All tests pass
- ✅ `go vet ./...` - No issues

**Files Changed:**
- Created: `providers/routerai/routerai.go`, `providers/routerai/types.go`, `providers/routerai/routerai_test.go`
- Modified: `config/model_type.go`, `factory/factory.go`, `factory/interface/factory.go`, `interfaces/providers.go`, `retrievers/qdrant.go`, `retrievers/sqlite.go`
- Deleted: `providers/openai/` (entire directory), `providers/ollama/` (entire directory)
- Updated: `go.mod` (removed langchaingo and related dependencies)

### Blockers 🚫
None currently.

### Technical Notes 📝

**Langchaingo Usage Found:**
| File | Packages | Complexity to Replace |
|------|----------|----------------------|
| `providers/openai/openai.go` | `llms`, `llms/openai` | 🟢 Easy - standard HTTP |
| `providers/ollama/ollama.go` | `llms`, `llms/ollama` | 🟢 Easy - same as OpenAI |
| `retrievers/qdrant.go` | `embeddings`, `schema`, `vectorstores`, `vectorstores/qdrant` | 🟡 Medium - needs Qdrant HTTP client |
| `retrievers/sqlite.go` | `embeddings`, `schema` | 🟢 Easy - stub implementation |
| `factory/factory.go` | `embeddings` | 🟢 Easy - update types |

**RouterAI.ru API:**
- Base URL: `https://routerai.ru/v1` (or configured)
- OpenAI-compatible endpoints
- Single API key authentication
- Models: configured in config.yml

**Implementation Plan:**
```go
// providers/routerai/routerai.go
// Just 2 methods needed:
// 1. GenerateResponse() -> POST /v1/chat/completions
// 2. CreateEmbedding() -> POST /v1/embeddings
```

---

## 📋 Backlog

### Technical Debt
- [ ] Make timeout configurable (currently `60 * time.Second` hardcoded)
- [x] Fix `factory.CreateRetriever()` - has undefined variables ✅ **FIXED** - Updated signature to use interfaces.VectorEmbeddingProvider
- [ ] Add bounds checking for string slicing in Telegram handlers

### Future Improvements
- [ ] Add retry logic for LLM calls
- [ ] Implement proper error handling with custom error types
- [ ] Add metrics/logging for LLM API calls

---

## 📚 Related Documents
- [Architecture Decision: Remove langchaingo](./docs/agents/decisions/001-remove-langchaingo.md)
- [Architecture Decision: RouterAI Provider](./docs/agents/decisions/002-routerai-provider.md)
- [Main Project Documentation](./AGENTS.md)

---

## Session History

| Date | Action | Result |
|------|--------|--------|
| 2026-02-03 | Evaluated langchaingo removal | Feasible, 2-3 days effort |
| 2026-02-03 | Selected RouterAI.ru as provider | Simplifies architecture |
| 2026-02-03 | Created migration documentation | Ready to start implementation |
| 2026-02-03 | **Completed full migration** | All tasks done, tests pass |

---

**Migration Status: ✅ COMPLETE**

The migration from langchaingo to direct HTTP calls with RouterAI.ru has been successfully completed. All planned tasks have been executed, all tests pass, and the codebase is now cleaner with ~53 fewer dependencies.
