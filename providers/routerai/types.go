package routerai

// ChatCompletionRequest represents the request body for chat completions
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents the response from chat completions
type ChatCompletionResponse struct {
	Choices []Choice `json:"choices"`
	Error   *Error   `json:"error,omitempty"`
}

// Choice represents a single completion choice
type Choice struct {
	Message Message `json:"message"`
}

// Error represents an API error response
type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// EmbeddingRequest represents the request body for embeddings
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse represents the response from embeddings
type EmbeddingResponse struct {
	Data  []EmbeddingData `json:"data"`
	Error *Error          `json:"error,omitempty"`
}

// EmbeddingData represents a single embedding result
type EmbeddingData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}
