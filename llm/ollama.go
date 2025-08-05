package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"

	"go-llm-rpggamemaster/clients"
	"go-llm-rpggamemaster/retrievers"
)

const (
	ollamaTimeout = 60 * time.Second
)

type OllamaClient struct {
	llmClient       *clients.OllamaLLMClient
	embeddingClient *clients.OllamaEmbeddingClient
	retriever       *retrievers.QdrantRetriever
	systemPrompt    *SystemPrompt
	messageTemplate string
}

func NewOllamaClient(model string) *OllamaClient {
	// Create LLM client
	llmClient, err := clients.NewOllamaLLMClient(model)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create Ollama client")
		panic(err)
	}

	// Create embedding client
	embeddingClient, err := clients.NewOllamaEmbeddingClient()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create embedding client")
		panic(err)
	}

	// Create retriever using the embedder from embedding client
	retriever, err := retrievers.NewQdrantRetriever(embeddingClient.GetEmbedder())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create qdrant retriever")
		panic(err)
	}

	return &OllamaClient{
		llmClient:       llmClient,
		embeddingClient: embeddingClient,
		retriever:       retriever,
		// Дефолтные настройки внутри клиента
		systemPrompt: &SystemPrompt{
			Role:    "system",
			Content: "Ты чат-бот дворецкий. Тебя зовут Игнациус. Отвечай вежливо и профессионально.",
		},
		messageTemplate: "{{.system}}\n\nВопрос: {{.question}}",
	}
}

// Методы для настройки (вызываются из main.go только для конфигурации)
func (c *OllamaClient) SetSystemPrompt(role, content string) {
	c.systemPrompt = &SystemPrompt{
		Role:    role,
		Content: content,
	}
}

func (c *OllamaClient) SetMessageTemplate(template string) {
	c.messageTemplate = template
}

// Применение шаблона к сообщению
func (c *OllamaClient) applyTemplate(userMessage string) string {
	if c.messageTemplate == "" {
		return userMessage
	}

	template := c.messageTemplate

	// Простая замена плейсхолдеров
	if c.systemPrompt != nil {
		template = strings.ReplaceAll(template, "{{.system}}", c.systemPrompt.Content)
	}
	template = strings.ReplaceAll(template, "{{.question}}", userMessage)

	return template
}

// Формирование полного промпта с системным сообщением
func (c *OllamaClient) buildFullPrompt(userMessage string) string {
	var fullPrompt strings.Builder

	// Добавляем системное сообщение, если оно есть
	if c.systemPrompt != nil && c.systemPrompt.Content != "" {
		fullPrompt.WriteString("Системные инструкции: ")
		fullPrompt.WriteString(c.systemPrompt.Content)
		fullPrompt.WriteString("\n\n")
	}

	// Применяем шаблон к пользовательскому сообщению
	processedMessage := c.applyTemplate(userMessage)
	fullPrompt.WriteString(processedMessage)

	return fullPrompt.String()
}

// Сохранение любого message в Qdrant с метаданными роли
func (c *OllamaClient) storeMessageInQdrant(ctx context.Context, role, content string) error {
	doc := schema.Document{
		PageContent: content,
		Metadata:    map[string]interface{}{"role": role, "ts": time.Now().Unix()},
	}
	return c.retriever.AddDocuments(ctx, []schema.Document{doc})
}

// Обычный генератор ответа: принимает вопрос, кидает в LLM, сохраняет и вопрос, и ответ в Qdrant
func (c *OllamaClient) GenerateSimpleResponse(ctx context.Context, userMessage string) (string, error) {
	return c.GenerateResponse(ctx, []Message{{Role: "user", Content: userMessage}}, 0.7, 0)
}

// GenerateResponse Для совместимости с интерфейсом. Теперь messages — слайс, можно сразу сохранить цепочку.
func (c *OllamaClient) GenerateResponse(ctx context.Context, messages []Message, temperature float64, maxTokens int) (string, error) {
	// Найдем самое свежее user-сообщение
	var userMsg string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			userMsg = messages[i].Content
			break
		}
	}
	if strings.TrimSpace(userMsg) == "" {
		return "", errors.New("no user message provided")
	}

	ctx, cancel := context.WithTimeout(ctx, ollamaTimeout)
	defer cancel()
	var opts []llms.CallOption
	if temperature != 0 {
		opts = append(opts, llms.WithTemperature(temperature))
	}
	if maxTokens > 0 {
		opts = append(opts, llms.WithMaxTokens(maxTokens))
	}

	// Добавить входящее сообщение игрока в вектор
	if err := c.storeMessageInQdrant(ctx, "user", userMsg); err != nil {
		log.Err(err).Msg("failed to store user message in qdrant")
	}

	// ИСПРАВЛЕНО: Используем buildFullPrompt вместо прямого userMsg
	fullPrompt := c.buildFullPrompt(userMsg)
	resp, err := c.llmClient.Call(ctx, fullPrompt, opts...)
	if err != nil {
		return "", fmt.Errorf("error from langchaingo Ollama: %w", err)
	}
	respStr := strings.TrimSpace(resp)
	if respStr == "" {
		return "", errors.New("empty response from Ollama/langchain")
	}

	// Записываем ответ LLM в Qdrant
	if err := c.storeMessageInQdrant(ctx, "assistant", respStr); err != nil {
		log.Err(err).Msg("failed to store assistant response in qdrant")
	}

	return respStr, nil
}

// Новый метод, если нужно формировать ответ на базе retrieval chain (поиск по истории в Qdrant)
func (c *OllamaClient) GenerateVectorResponse(ctx context.Context, userMessage string, temperature float64, maxTokens int) (string, error) {
	// 1. Настройка опций для цепочки
	var opts []chains.ChainCallOption
	if temperature > 0 {
		opts = append(opts, chains.WithTemperature(temperature))
	}
	if maxTokens > 0 {
		opts = append(opts, chains.WithMaxTokens(maxTokens))
	}
	ctx, cancel := context.WithTimeout(ctx, ollamaTimeout)
	defer cancel()

	// 2. Создание retriever для получения документов из Qdrant
	// We need to implement a wrapper for our retriever to make it compatible
	// For now, we'll create a compatible retriever manually
	docs, err := c.retriever.GetRelevantDocuments(ctx, userMessage)
	if err != nil {
		return "", err
	}

	// 3. Создание шаблона промпта, который включает системное сообщение и контекст
	// Этот шаблон будет использоваться для "сборки" финального запроса к LLM
	ragTemplate := `{{.system_prompt}}

Используй следующий контекст из истории диалога, чтобы ответить на вопрос.

Контекст:
{{.chatContext}}

Вопрос: {{.question}}

Ответ:`

	// Подставляем наш системный промпт в шаблон
	fullRagPromptStr := strings.Replace(ragTemplate, "{{.system_prompt}}", c.systemPrompt.Content, 1)

	// Creating a wrapper to make our client compatible with llms.Model
	llmWrapper := &llmWrapper{client: c.llmClient}

	// Создаем объект шаблона
	prompt := prompts.NewPromptTemplate(
		fullRagPromptStr,
		[]string{"chatContext", "question"},
	)

	// Format chatContext from documents
	var chatContext strings.Builder
	for _, doc := range docs {
		chatContext.WriteString(doc.PageContent)
		chatContext.WriteString("\n")
	}

	// 4. Создаем цепочку для работы с LLM и нашим шаблоном
	llmChain := chains.NewLLMChain(llmWrapper, prompt)

	// 5. Вызываем цепочку напрямую без RetrievalQA
	result, err := chains.Call(ctx, llmChain, map[string]any{
		"context":  chatContext.String(),
		"question": userMessage,
	}, opts...)
	if err != nil {
		log.Err(err).Msg("ошибка при вызове цепочки langchaingo LLMChain")
		return "", err
	}

	// 6. Извлекаем результат
	text, ok := result["text"].(string) // Стандартный выходной ключ
	if !ok || strings.TrimSpace(text) == "" {
		return "", errors.New("нет ответа от векторной цепочки")
	}

	// Сохраняем сообщения в Qdrant
	if err := c.storeMessageInQdrant(ctx, "user", userMessage); err != nil {
		log.Err(err).Msg("не удалось сохранить сообщение пользователя в qdrant (vector)")
	}
	if err := c.storeMessageInQdrant(ctx, "assistant", text); err != nil {
		log.Err(err).Msg("не удалось сохранить ответ ассистента в qdrant (vector)")
	}
	return text, nil
}

// Wrapper to make our client compatible with llms.Model
type llmWrapper struct {
	client *clients.OllamaLLMClient
}

func (w *llmWrapper) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return w.client.Call(ctx, prompt, options...)
}

func (w *llmWrapper) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	// For simplicity, we'll convert the first text message to a prompt
	if len(messages) > 0 && len(messages[0].Parts) > 0 {
		if textPart, ok := messages[0].Parts[0].(llms.TextContent); ok {
			response, err := w.client.Call(ctx, textPart.Text, options...)
			if err != nil {
				return nil, err
			}
			return &llms.ContentResponse{
				Choices: []*llms.ContentChoice{
					{
						Content: response,
					},
				},
			}, nil
		}
	}
	return nil, errors.New("unsupported message format")
}

func NewRPGOllamaClient(model string) *OllamaClient {
	client := NewOllamaClient(model)
	client.SetSystemPrompt("system", "Ты опытный мастер RPG игр по имени Игнациус.")
	client.SetMessageTemplate("Мастер: {{.system}}\nИгрок: {{.question}}\nОтвет:")
	return client
}
