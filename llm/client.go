package llm

import "context"

// LLMClient представляет интерфейс для взаимодействия с языковыми моделями
type LLMClient interface {
	GenerateResponse(ctx context.Context, messages []Message, temperature float64, maxTokens int) (string, error)
	GenerateSimpleResponse(ctx context.Context, userMessage string) (string, error)
}
