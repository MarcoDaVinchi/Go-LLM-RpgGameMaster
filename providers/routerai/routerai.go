package routerai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-llm-rpggamemaster/interfaces"
)

const (
	defaultTimeout = 60 * time.Second
	defaultBaseURL = "https://routerai.ru/v1"
)

// RouterAIProvider implements both InferenceProvider and VectorEmbeddingProvider
type RouterAIProvider struct {
	model   string
	apiKey  string
	baseURL string
	client  *http.Client
}

var (
	_ interfaces.InferenceProvider       = (*RouterAIProvider)(nil)
	_ interfaces.VectorEmbeddingProvider = (*RouterAIProvider)(nil)
)

func NewRouterAIProvider(model, apiKey, baseURL string) (*RouterAIProvider, error) {
	if model == "" {
		return nil, fmt.Errorf("model name is required")
	}
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &RouterAIProvider{
		model:   model,
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: defaultTimeout},
	}, nil
}

func (p *RouterAIProvider) GenerateResponse(ctx context.Context, messages []interfaces.Message, temperature float64, maxTokens int) (string, error) {
	var reqMessages []Message
	for _, m := range messages {
		reqMessages = append(reqMessages, Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	reqBody := ChatCompletionRequest{
		Model:       p.model,
		Messages:    reqMessages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var chatResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (p *RouterAIProvider) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := EmbeddingRequest{
		Model: p.model,
		Input: texts,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/embeddings", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var embedResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if embedResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", embedResp.Error.Message)
	}

	result := make([][]float32, len(texts))
	for _, data := range embedResp.Data {
		if data.Index < len(result) {
			result[data.Index] = data.Embedding
		}
	}

	return result, nil
}

func (p *RouterAIProvider) Name() string {
	return "routerai"
}
