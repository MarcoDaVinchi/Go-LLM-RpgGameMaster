package llm

import "context"

// LLMClient представляет интерфейс для взаимодействия с языковыми моделями
type LLMClient interface {
	GenerateResponse(ctx context.Context, messages []Message, temperature float64, maxTokens int) (string, error)
	GenerateSimpleResponse(ctx context.Context, userMessage string) (string, error)
}

// Message представляет сообщение в диалоге
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type SystemPrompt struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
