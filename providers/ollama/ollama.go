package ollama

import (
	"context"
	"fmt"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"go-llm-rpggamemaster/interfaces"
)

const (
	defaultOllamaModel = "llama3.1"
	ollamaTimeout      = 60 * time.Second
)

// OllamaProvider implements the LLMProvider interface for Ollama
type OllamaProvider struct {
	model string
	llm   *ollama.LLM
}

// NewOllamaProvider creates a new Ollama provider instance
func NewOllamaProvider(model, baseURL string) (interfaces.LLMProvider, error) {
	if model == "" {
		model = defaultOllamaModel
	}

	var opts []ollama.Option
	opts = append(opts, ollama.WithModel(model))
	if baseURL != "" {
		opts = append(opts, ollama.WithServerURL(baseURL))
	}

	llm, err := ollama.New(opts...)
	if err != nil {
		return nil, err
	}

	return &OllamaProvider{
		model: model,
		llm:   llm,
	}, nil
}

// GenerateResponse generates a response using the Ollama provider
func (p *OllamaProvider) GenerateResponse(ctx context.Context, messages []interfaces.Message, temperature float64, maxTokens int) (string, error) {
	// Convert interfaces.Message to llms.MessageContent
	var llmMessages []llms.MessageContent
	for _, msg := range messages {
		llmMessages = append(llmMessages, llms.MessageContent{
			Role:  llms.ChatMessageType(msg.Role),
			Parts: []llms.ContentPart{llms.TextContent{Text: msg.Content}},
		})
	}

	// Set up options
	var opts []llms.CallOption
	if temperature != 0 {
		opts = append(opts, llms.WithTemperature(temperature))
	}
	if maxTokens > 0 {
		opts = append(opts, llms.WithMaxTokens(maxTokens))
	}

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, ollamaTimeout)
	defer cancel()

	// Generate response
	response, err := p.llm.GenerateContent(ctx, llmMessages, opts...)
	if err != nil {
		return "", fmt.Errorf("error generating response: %w", err)
	}

	if len(response.Choices) == 0 || response.Choices[0] == nil {
		return "", fmt.Errorf("no response generated")
	}

	content := response.Choices[0].Content
	if content == "" {
		return "", fmt.Errorf("empty response generated")
	}

	return content, nil
}

// GenerateSimpleResponse generates a simple response using the Ollama provider
func (p *OllamaProvider) GenerateSimpleResponse(ctx context.Context, userMessage string) (string, error) {
	messages := []interfaces.Message{
		{Role: "user", Content: userMessage},
	}
	return p.GenerateResponse(ctx, messages, 0.7, 0)
}

// Name returns the name of the provider
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// OllamaEmbeddingProvider implements the EmbeddingProvider interface for Ollama
type OllamaEmbeddingProvider struct {
	model    string
	embedder *ollama.LLM
}

// NewOllamaEmbeddingProvider creates a new Ollama embedding provider instance
func NewOllamaEmbeddingProvider(model, baseURL string) (interfaces.EmbeddingProvider, error) {
	if model == "" {
		model = "nomic-embed-text"
	}

	var opts []ollama.Option
	opts = append(opts, ollama.WithModel(model))
	if baseURL != "" {
		opts = append(opts, ollama.WithServerURL(baseURL))
	}

	embedder, err := ollama.New(opts...)
	if err != nil {
		return nil, err
	}

	return &OllamaEmbeddingProvider{
		model:    model,
		embedder: embedder,
	}, nil
}

// EmbedDocuments embeds a list of documents using the Ollama provider
func (p *OllamaEmbeddingProvider) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	ctx, cancel := context.WithTimeout(ctx, ollamaTimeout)
	defer cancel()

	return p.embedder.CreateEmbedding(ctx, texts)
}

// Name returns the name of the provider
func (p *OllamaEmbeddingProvider) Name() string {
	return "ollama"
}
