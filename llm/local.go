package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	localModelURL = "http://localhost:1234/v1/chat/completions"
	localTimeout  = 60 * time.Second
)

// LocalModelClient представляет клиент для взаимодействия с локальной моделью
type LocalModelClient struct {
	model      string
	httpClient *http.Client
}

// NewLocalModelClient создает новый клиент для локальной модели
func NewLocalModelClient(model string) *LocalModelClient {
	if model == "" {
		model = defaultModel
	}

	return &LocalModelClient{
		model: model,
		httpClient: &http.Client{
			Timeout: localTimeout,
		},
	}
}

// GenerateResponse генерирует ответ от локальной модели
func (c *LocalModelClient) GenerateResponse(ctx context.Context, messages []Message, temperature float64, maxTokens int) (string, error) {
	request := ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, localModelURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Не требуется заголовок Authorization для локальной модели

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to local model: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("local model API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var response ChatCompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", errors.New("no choices in response from local model")
	}

	return response.Choices[0].Message.Content, nil
}

// GenerateSimpleResponse - удобный метод для генерации ответа с одним сообщением пользователя
func (c *LocalModelClient) GenerateSimpleResponse(ctx context.Context, userMessage string) (string, error) {
	messages := []Message{
		{
			Role:    "user",
			Content: userMessage,
		},
	}
	return c.GenerateResponse(ctx, messages, 0.7, 0) // Стандартная температура и без ограничения токенов
}
