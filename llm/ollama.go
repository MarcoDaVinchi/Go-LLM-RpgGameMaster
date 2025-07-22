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

type OllamaClient struct {
	model    string
	llm      *ollama.LLM
	embedder *embeddings.EmbedderImpl
	qstore   qdrant.Store
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
	}
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

// Для совместимости с интерфейсом. Теперь messages — слайс, можно сразу сохранить цепочку.
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

	// Сначала получаем ответ (как раньше)
	resp, err := c.llm.Call(ctx, userMsg, opts...)
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
	var opts []chains.ChainCallOption
	if temperature > 0 {
		opts = append(opts, chains.WithTemperature(temperature))
	}
	if maxTokens > 0 {
		opts = append(opts, chains.WithMaxTokens(maxTokens))
	}
	ctx, cancel := context.WithTimeout(ctx, ollamaTimeout)
	defer cancel()

	retriever := vectorstores.ToRetriever(c.qstore, 10)
	//template := `Ты чат бот дворецкий. Тебя зовут Игнациус. К тебе обратились с сообщением "{{.question}}". Используй информацию из базы для ответа.`
	chain := chains.NewRetrievalQAFromLLM(c.llm, retriever)
	result, err := chain.Call(ctx, map[string]interface{}{"question": userMessage}, opts...)
	if err != nil {
		log.Err(err).Msg("langchaingo RetrievalQAChain error")
	}
	text, ok := result["text"].(string)
	if !ok || strings.TrimSpace(text) == "" {
		return "", errors.New("нет ответа от векторной цепочки")
	}
	// Дополнительно: можно тоже сохранять userMessage и результат chain в Qdrant
	if err := c.storeMessageInQdrant(ctx, "user", userMessage); err != nil {
		log.Err(err).Msg("failed to store user message in qdrant (vector)")
	}
	if err := c.storeMessageInQdrant(ctx, "assistant", text); err != nil {
		log.Err(err).Msg("failed to store assistant response in qdrant (vector)")
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
