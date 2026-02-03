# Active TODOs & Migration Plans

**Last Updated:** 2026-02-03  
**Current Focus:** Remove langchaingo dependency, migrate to RouterAI.ru

---

## 🚀 [MIGRATION-001] Remove langchaingo → Direct HTTP

**Status:** Planning Complete, Ready to Start  
**Started:** 2026-02-03  
**Estimated Effort:** 2-3 days  
**Priority:** High

### Context
Migration from langchaingo library to direct HTTP calls. Langchaingo hides simple HTTP interactions behind heavy abstraction layer with weak maintainability. RouterAI.ru (OpenAI-compatible API) will be the single LLM provider.

### Completed ✅
- [x] Assessed langchaingo usage across codebase
- [x] Evaluated replacement complexity (LOW)
- [x] Selected RouterAI.ru as unified provider (OpenAI-compatible)
- [x] Created migration documentation

### In Progress 🔄
- [ ] Create `providers/routerai/` package
- [ ] Implement HTTP client for RouterAI.ru
- [ ] Add chat completion endpoint (`POST /v1/chat/completions`)
- [ ] Add embeddings endpoint (`POST /v1/embeddings`)

### Pending 📋
- [ ] Update factory to use RouterAI provider
- [ ] Remove `providers/openai/` (replaced by routerai)
- [ ] Remove `providers/ollama/` (replaced by routerai)
- [ ] Refactor retrievers to remove langchaingo dependencies
- [ ] Update `go.mod` - remove langchaingo
- [ ] Run `go mod tidy`
- [ ] Update tests
- [ ] Verify all tests pass
- [ ] Update AGENTS.md with new architecture

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
- [ ] Fix `factory.CreateRetriever()` - has undefined variables
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

---

**Next Session Start Here:**
1. Read `docs/agents/decisions/001-remove-langchaingo.md`
2. Read `docs/agents/decisions/002-routerai-provider.md`
3. Start with `providers/routerai/routerai.go` implementation
4. Check off items in "In Progress" section above
