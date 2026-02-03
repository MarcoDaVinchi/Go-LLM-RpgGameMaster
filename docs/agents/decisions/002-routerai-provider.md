# ADR 002: RouterAI.ru as Unified LLM Provider

## Status
**Accepted** (2026-02-03)

**Decision Owner:** @Sisyphus (AI Agent)  
**Stakeholders:** Project maintainers

---

## Context

As part of the migration away from langchaingo (see [ADR 001](./001-remove-langchaingo.md)), we need to select a unified LLM provider strategy.

Previously, the project supported multiple backends:
- **OpenAI** - Cloud API, requires API key, proprietary
- **Ollama** - Local deployment, open-source models

Supporting multiple providers adds complexity:
- Different HTTP APIs (though both have OpenAI-compatible modes)
- Different authentication methods
- Different model names and capabilities
- More configuration options

### Requirements

1. **OpenAI API compatibility** - Standard interface, well-documented
2. **Multiple model support** - Access to various LLM models
3. **Cost-effective** - Competitive pricing
4. **Reliable** - Good uptime and performance
5. **Simple configuration** - Single API key, standard endpoints

---

## Decision

**Use RouterAI.ru as the unified LLM provider.**

### Provider Details

| Attribute | Value |
|-----------|-------|
| **Service** | RouterAI.ru |
| **API Type** | OpenAI-compatible |
| **Base URL** | `https://routerai.ru/v1` (configurable) |
| **Authentication** | API Key (Bearer token) |
| **Endpoints** | `/v1/chat/completions`, `/v1/embeddings` |

### Rationale

1. **OpenAI Compatibility:** Standard API means:
   - Well-documented endpoints
   - Existing client libraries work (if needed later)
   - Easy to switch between compatible providers

2. **Simplified Architecture:**
   - Single provider implementation
   - Single authentication method
   - Consistent error handling
   - Unified configuration

3. **Model Variety:**
   - Access to multiple models through single API
   - Mix of proprietary and open-source models
   - See: https://routerai.ru/models

4. **Cost Optimization:**
   - Competitive pricing compared to direct OpenAI
   - Can select models based on price/performance

---

## Consequences

### Positive ➕

- **Simpler codebase:** One provider instead of two
- **Standard API:** OpenAI-compatible means familiar interface
- **Easier testing:** Single API to mock
- **Consistent behavior:** Same request/response format for all models
- **Reduced complexity:** Factory logic simplified
- **Better maintainability:** Less code to maintain

### Negative ➖

- **Vendor lock-in:** Dependent on RouterAI.ru availability
- **Internet required:** No local deployment option (Ollama removed)
- **API key management:** Single point of failure for credentials
- **Migration effort:** Need to remove Ollama support

### Mitigations

- **API compatibility:** Easy to switch to another OpenAI-compatible provider
- **Configuration:** URL and API key are configurable (can point to OpenAI directly)
- **Abstraction:** Keep `InferenceProvider` interface for future flexibility

---

## Implementation

### Configuration

```yaml
# config.yml
inference_model:
  url: "${ROUTERAI_URL:https://routerai.ru/v1}"
  type: "routerai"
  name: "moonshotai/kimi-k2.5"
  api_key: "${ROUTERAI_API_KEY}"

embedding_model:
  url: "${ROUTERAI_URL:https://routerai.ru/v1}"
  type: "routerai"
  name: "text-embedding-ada-002"
  api_key: "${ROUTERAI_API_KEY}"
```

### Environment Variables

```bash
ROUTERAI_API_KEY=your_api_key_here
ROUTERAI_URL=https://routerai.ru/v1  # Optional, has default
```

### Code Structure

```
providers/
├── routerai/
│   ├── routerai.go          # Main provider implementation
│   ├── routerai_test.go     # Tests
│   └── types.go             # Request/response structs
└── [openai/ - REMOVED]
└── [ollama/ - REMOVED]
```

### Model Selection

RouterAI.ru supports many models. Recommended for this project:

| Use Case | Model | Notes |
|----------|-------|-------|
| **Primary LLM** | `moonshotai/kimi-k2.5` | Good balance of quality/price |
| **Fast tasks** | `z-ai/glm-4.7-flash` | Cheapest option |
| **Embeddings** | `text-embedding-ada-002` | Standard OpenAI embeddings |

---

## Migration Path

### From OpenAI

1. Change URL from `https://api.openai.com/v1` to `https://routerai.ru/v1`
2. Update API key
3. Select appropriate model from RouterAI catalog

### From Ollama

1. Remove local Ollama dependency
2. Configure RouterAI.ru URL and API key
3. Select cloud-based model equivalent

---

## Related Decisions

- [ADR 001: Remove langchaingo](./001-remove-langchaingo.md) - Parent decision enabling this simplification

---

## References

- RouterAI.ru: https://routerai.ru
- Available Models: https://routerai.ru/models
- OpenAI API Compatibility: https://platform.openai.com/docs/api-reference
