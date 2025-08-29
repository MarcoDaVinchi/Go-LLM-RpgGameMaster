package clients

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

const (
	inferenceModelDefaultTimeout = 60 * time.Second
)

type InferenceClient struct {
	model string
	llm   *openai.LLM
}

func NewInferenceClient(model string) (*InferenceClient, error) {
	if model == "" {
		log.Fatal().Msg("inference model is not set")
	}

	llm, err := openai.New(openai.WithModel(model))
	if err != nil {
		return nil, err
	}

	return &InferenceClient{
		model: model,
		llm:   llm,
	}, nil
}

func (c *InferenceClient) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	if _, ok := ctx.Deadline(); !ok {
		_, cancel := context.WithTimeout(ctx, inferenceModelDefaultTimeout)
		defer cancel()
	}

	return c.llm.Call(ctx, prompt, options...)
}
