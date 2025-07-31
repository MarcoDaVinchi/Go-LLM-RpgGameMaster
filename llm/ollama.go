package llm

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/qdrant"
)

const (
	defaultOllamaModel    = "llama3.1"
	defaultEmbeddingModel = "nomic-embed-text"
	ollamaTimeout         = 60 * time.Second
	defaultCollection     = "game_collection"
)

type SystemPrompt struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaClient struct {
	model           string
	llm             *ollama.LLM
	embedder        *embeddings.EmbedderImpl
	qstore          qdrant.Store
	systemPrompt    *SystemPrompt
	messageTemplate string
}

func NewOllamaClient(model string) *OllamaClient {
	if model == "" {
		model = defaultOllamaModel
	}
	llm, err := ollama.New(ollama.WithModel(model))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create Ollama client")
		panic(err)
	}
	embedLlm, err := ollama.New(ollama.WithModel(defaultEmbeddingModel))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create embedding client")
		panic(err)
	}
	embedder, err := embeddings.NewEmbedder(embedLlm)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create embedder")
		panic(err)
	}
	qstore, err := initQdrantRetriever(embedder)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create qdrant retriever")
		panic(err)
	}
	return &OllamaClient{
		model:    model,
		llm:      llm,
		embedder: embedder,
		qstore:   qstore,
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
	_, err := c.qstore.AddDocuments(ctx, []schema.Document{doc})
	return err
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
	resp, err := c.llm.Call(ctx, fullPrompt, opts...)
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
	retriever := vectorstores.ToRetriever(c.qstore, 10)

	// 3. Создание шаблона промпта, который включает системное сообщение и контекст
	// Этот шаблон будет использоваться для "сборки" финального запроса к LLM
	ragTemplate := `{{.system_prompt}}

Используй следующий контекст из истории диалога, чтобы ответить на вопрос.

Контекст:
{{.context}}

Вопрос: {{.question}}

Ответ:`

	// Подставляем наш системный промпт в шаблон
	fullRagPromptStr := strings.Replace(ragTemplate, "{{.system_prompt}}", c.systemPrompt.Content, 1)

	// Создаем объект шаблона
	prompt := prompts.NewPromptTemplate(
		fullRagPromptStr,
		[]string{"context", "question"},
	)

	// 4. Создаем цепочку для работы с LLM и нашим шаблоном
	llmChain := chains.NewLLMChain(c.llm, prompt)

	// 5. Создаем цепочку, которая "наполняет" (stuffs) документы в LLMChain
	stuffChain := chains.NewStuffDocuments(llmChain)

	// 6. Собираем основную RAG цепочку из "наполнителя" и retriever
	qaChain := chains.NewRetrievalQA(stuffChain, retriever)

	// 7. Вызываем итоговую цепочку
	result, err := chains.Call(ctx, qaChain, map[string]any{
		"query": userMessage, // Стандартный входной ключ для RetrievalQA
	}, opts...)
	if err != nil {
		log.Err(err).Msg("ошибка при вызове цепочки langchaingo RetrievalQA")
		return "", err
	}

	// 8. Извлекаем результат
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

func initQdrantRetriever(embedder *embeddings.EmbedderImpl) (qdrant.Store, error) {
	var qdrantUrl, err = url.Parse("http://localhost:6333")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse qdrant url")
		panic(err)
	}
	return qdrant.New(
		qdrant.WithURL(*qdrantUrl),
		qdrant.WithCollectionName(defaultCollection),
		qdrant.WithEmbedder(embedder),
	)
}

func NewRPGOllamaClient(model string) *OllamaClient {
	client := NewOllamaClient(model)
	client.SetSystemPrompt("system", "Ты опытный мастер RPG игр по имени Игнациус.")
	client.SetMessageTemplate("Мастер: {{.system}}\nИгрок: {{.question}}\nОтвет:")
	return client
}
