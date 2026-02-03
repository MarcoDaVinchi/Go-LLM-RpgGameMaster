# Known Issues & Technical Debt

**Last Updated:** 2026-02-03

---

## 🐛 Known Issues

### Critical

None currently.

### High Priority

#### 1. Factory.CreateRetriever() - Incomplete Implementation
**Location:** `factory/factory.go:62-78`  
**Status:** Broken code  
**Impact:** Retriever creation will fail at runtime

```go
// Problem: undefined variables 'embedder' and 'retrieverType'
func (f *providerFactory) CreateRetriever() (retrievers.Retriever, error) {
    embedderImpl, ok := embedder.(*embeddings.EmbedderImpl)  // 'embedder' undefined
    // ...
    switch retrieverType {  // 'retrieverType' undefined
```

**Fix Required:**
- Pass embedder and retrieverType as parameters or from config
- Update factory interface if needed

### Medium Priority

#### 2. String Slicing Without Bounds Checking
**Location:** Telegram handlers in `main.go`  
**Pattern:** `Text[len("/cmd"):]`

**Risk:** Panic if message is shorter than command prefix  
**Fix:** Add length check before slicing

```go
// Unsafe:
content := message.Text[len("/cmd"):]

// Safe:
const cmdPrefix = "/cmd"
if len(message.Text) > len(cmdPrefix) {
    content = message.Text[len(cmdPrefix):]
}
```

### Low Priority

#### 3. No Retry Logic for LLM Calls
**Location:** All providers  
**Impact:** Transient failures cause immediate errors  
**Fix:** Add exponential backoff retry mechanism

---

## 📊 Technical Debt

### Hardcoded Values

| Value | Location | Should Be |
|-------|----------|-----------|
| `60 * time.Second` | `providers/*/*.go` | Configurable timeout |
| `gpt-3.5-turbo` | `providers/openai/openai.go` | Config or constant |
| `llama3.1` | `providers/ollama/ollama.go` | Config or constant |
| `nomic-embed-text` | `providers/ollama/ollama.go` | Config or constant |
| `text-embedding-ada-002` | `providers/openai/openai.go` | Config or constant |
| `game_collection` | `retrievers/qdrant.go` | Configurable |
| Temperature `0.7` | `GenerateSimpleResponse()` | Parameter with default |

### Testing Issues

| Issue | Impact | Solution |
|-------|--------|----------|
| Tests require running services | Can't run in CI | Add mocks or test containers |
| No unit tests for HTTP clients | Low coverage | Add with httptest |
| Integration tests in main package | Organization | Move to _test packages |

### Dependency Concerns

| Dependency | Concern | Action |
|------------|---------|--------|
| `github.com/tmc/langchaingo` | Weak maintenance | **IN PROGRESS: Remove** |
| `github.com/testcontainers/testcontainers-go` | Heavy for tests | Evaluate if needed |

---

## 🔄 Current Migration

### In Progress: Remove langchaingo

See [AGENTS-TODO.md](../../AGENTS-TODO.md) for detailed plan.

**Related:**
- [ADR 001: Remove langchaingo](../decisions/001-remove-langchaingo.md)
- [ADR 002: RouterAI Provider](../decisions/002-routerai-provider.md)

---

## 📝 Notes for Future Agents

### Before Making Changes

1. Check this file for known issues in files you're modifying
2. Update status if you fix an issue
3. Add new issues discovered during work

### When Adding New Providers

- Avoid langchaingo patterns (we're removing it)
- Use direct HTTP or official lightweight SDKs
- Keep interfaces simple

### When Modifying Retrievers

- Fix `factory.CreateRetriever()` first (it's broken)
- Ensure embedder is properly initialized
- Test with both Qdrant and SQLite

---

## History

| Date | Change | Author |
|------|--------|--------|
| 2026-02-03 | Initial documentation | @Sisyphus |
| 2026-02-03 | Added migration context | @Sisyphus |
