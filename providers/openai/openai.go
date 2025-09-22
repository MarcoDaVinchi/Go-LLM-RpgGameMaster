package openai

import (
	"context"
	"fmt"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"go-llm-rpggamemaster/interfaces"
)

const (
	defaultOpenAIModel = "gpt-3.5-turbo"
	openAITimeout      = 60 * time.Second
)

// OpenAIProvider implements the InferenceProvider interface for OpenAI
type OpenAIProvider struct {
	model string
	llm   *openai.LLM
}

// NewOpenAIProvider creates a new OpenAI provider instance
func NewOpenAIProvider(model, apiKey, baseURL string) (interfaces.InferenceProvider, error) {
	if model == "" {
		model = defaultOpenAIModel
	}

	var opts []openai.Option
	if apiKey != "" {
		opts = append(opts, openai.WithToken(apiKey))
	}
	if baseURL != "" {
		opts = append(opts, openai.WithBaseURL(baseURL))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return nil, err
	}

	return &OpenAIProvider{
		model: model,
		llm:   llm,
	}, nil
}

// GenerateResponse generates a response using the OpenAI provider
func (p *OpenAIProvider) GenerateResponse(ctx context.Context, messages []interfaces.Message, temperature float64, maxTokens int) (string, error) {
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
	ctx, cancel := context.WithTimeout(ctx, openAITimeout)
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

// GenerateSimpleResponse generates a simple response using the OpenAI provider
func (p *OpenAIProvider) GenerateSimpleResponse(ctx context.Context, userMessage string) (string, error) {
	messages := []interfaces.Message{
		{Role: "user", Content: userMessage},
	}
	return p.GenerateResponse(ctx, messages, 0.7, 0)
}

// Name returns the name of the provider
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// OpenAIEmbeddingProvider implements the VectorEmbeddingProvider interface for OpenAI
type OpenAIEmbeddingProvider struct {
	model    string
	embedder *openai.LLM
}

// NewOpenAIEmbeddingProvider creates a new OpenAI embedding provider instance
func NewOpenAIEmbeddingProvider(model, apiKey, baseURL string) (interfaces.VectorEmbeddingProvider, error) {
	if model == "" {
		model = "text-embedding-ada-002"
	}

	var opts []openai.Option
	if apiKey != "" {
		opts = append(opts, openai.WithToken(apiKey))
	}
	if baseURL != "" {
		opts = append(opts, openai.WithBaseURL(baseURL))
	}

	embedder, err := openai.New(append(opts, openai.WithModel(model))...)
	if err != nil {
		return nil, err
	}

	return &OpenAIEmbeddingProvider{
		model:    model,
		embedder: embedder,
	}, nil
}

// EmbedDocuments embeds a list of documents using the OpenAI provider
func (p *OpenAIEmbeddingProvider) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	ctx, cancel := context.WithTimeout(ctx, openAITimeout)
	defer cancel()

	return p.embedder.CreateEmbedding(ctx, texts)
}

// Name returns the name of the provider
func (p *OpenAIEmbeddingProvider) Name() string {
	return "openai"
}
