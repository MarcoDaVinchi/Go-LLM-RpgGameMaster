package clients

import (
	"context"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

const (
	defaultOllamaModel = "llama3.1"
	ollamaTimeout      = 60 * time.Second
)

type LLMClient interface {
	Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error)
}

type OllamaLLMClient struct {
	model string
	llm   *ollama.LLM
}

func NewOllamaLLMClient(model string) (*OllamaLLMClient, error) {
	if model == "" {
		model = defaultOllamaModel
	}

	llm, err := ollama.New(ollama.WithModel(model))
	if err != nil {
		return nil, err
	}

	return &OllamaLLMClient{
		model: model,
		llm:   llm,
	}, nil
}

func (c *OllamaLLMClient) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, ollamaTimeout)
	defer cancel()

	return c.llm.Call(ctx, prompt, options...)
}
