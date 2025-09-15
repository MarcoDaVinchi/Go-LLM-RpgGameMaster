package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"go-llm-rpggamemaster/config"
	factory "go-llm-rpggamemaster/factory"
	"go-llm-rpggamemaster/interfaces"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var llmProvider interfaces.LLMProvider

// validateEnvironmentVariables checks for required environment variables based on configuration
func validateEnvironmentVariables(cfg *config.Config) error {
	// Check QDRANT_URL if qdrant retriever is configured
	for _, retriever := range cfg.Retrievers {
		if retriever.Name == "qdrant" {
			if qdrantURL := os.Getenv("QDRANT_URL"); qdrantURL == "" {
				return fmt.Errorf("QDRANT_URL environment variable is required when using qdrant retriever")
			}
			log.Info().Str("qdrant_url", os.Getenv("QDRANT_URL")).Msg("Qdrant URL configured")
		}
	}

	// Check OPENAI_API_KEY if openai provider is configured
	for _, provider := range cfg.LLMProviders {
		if provider.Name == "openai" {
			if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey == "" {
				return fmt.Errorf("OPENAI_API_KEY environment variable is required when using openai LLM provider")
			}
			log.Info().Msg("OpenAI API key configured for LLM provider")
		}
	}
	for _, provider := range cfg.EmbeddingProviders {
		if provider.Name == "openai" {
			if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey == "" {
				return fmt.Errorf("OPENAI_API_KEY environment variable is required when using openai embedding provider")
			}
			log.Info().Msg("OpenAI API key configured for embedding provider")
		}
	}

	// Check RPG_TELEGRAM_BOT_API_KEY (already checked later, but validate here for consistency)
	if token := os.Getenv("RPG_TELEGRAM_BOT_API_KEY"); token == "" {
		return fmt.Errorf("RPG_TELEGRAM_BOT_API_KEY environment variable is required")
	}
	log.Info().Msg("Telegram bot API key configured")

	// Log optional environment variables that may affect configuration
	if inferenceModel := os.Getenv("INFERENCE_MODEL"); inferenceModel != "" {
		log.Info().Str("inference_model", inferenceModel).Msg("Custom inference model configured")
	}
	if embeddingModel := os.Getenv("EMBEDDING_MODEL"); embeddingModel != "" {
		log.Info().Str("embedding_model", embeddingModel).Msg("Custom embedding model configured")
	}
	if inferenceURL := os.Getenv("INFERENCE_SERVER_URL"); inferenceURL != "" {
		log.Info().Str("inference_server_url", inferenceURL).Msg("Custom inference server URL configured")
	}
	if embeddingURL := os.Getenv("EMBEDDING_SERVER_URL"); embeddingURL != "" {
		log.Info().Str("embedding_server_url", embeddingURL).Msg("Custom embedding server URL configured")
	}

	return nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Load configuration
	err := godotenv.Load()
	if err != nil {
		log.Info().Msgf("Env file is missing %s", err)
	}
	profile := os.Getenv("PROFILE")
	if profile == "local" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	cfg, err := config.LoadConfigFromDefaultPath()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	// Log loaded configuration parameters for debugging
	log.Info().Msg("Configuration loaded successfully")
	log.Info().Str("default_llm", cfg.DefaultLLM).Msg("Default LLM provider")
	log.Info().Str("default_embedding", cfg.DefaultEmbedding).Msg("Default embedding provider")
	log.Info().Str("default_retriever", cfg.DefaultRetriever).Msg("Default retriever")

	for _, provider := range cfg.LLMProviders {
		log.Info().Str("provider", provider.Name).Str("model", provider.Model).Msg("LLM provider configured")
	}
	for _, provider := range cfg.EmbeddingProviders {
		log.Info().Str("provider", provider.Name).Str("model", provider.Model).Msg("Embedding provider configured")
	}
	for _, retriever := range cfg.Retrievers {
		log.Info().Str("retriever", retriever.Name).Str("db_path", retriever.DBPath).Str("collection", retriever.Collection).Msg("Retriever configured")
	}

	// Validate required environment variables based on configuration
	if err := validateEnvironmentVariables(cfg); err != nil {
		log.Fatal().Err(err).Msg("environment validation failed")
	}

	// Create provider factory
	providerFactory := factory.NewProviderFactory(cfg)

	// Create LLM provider
	llmProvider, err = providerFactory.CreateDefaultLLMProvider()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create LLM provider")
	}

	log.Info().Msgf("Using LLM provider: %s", llmProvider.Name())

	opts := []bot.Option{
		bot.WithDefaultHandler(gptHandler),
	}
	token := os.Getenv("RPG_TELEGRAM_BOT_API_KEY")
	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/echo", bot.MatchTypePrefix, echoHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/status", bot.MatchTypePrefix, userStatusHandler)
	b.Start(ctx)
}

func gptHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	prompt := strings.TrimSpace(update.Message.Text[len("/gpt"):])
	if prompt == "" {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Пожалуйста, укажите запрос после команды /gpt",
		})
		if err != nil {
			log.Err(err).Msg("failed to send message")
			return
		}
		return
	}

	response, err := llmProvider.GenerateSimpleResponse(ctx, prompt)
	if err != nil {
		log.Err(err).Msg("failed to get response from LLM provider")
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Ошибка обращения к %s: %s", llmProvider.Name(), err.Error()),
		})
		if err != nil {
			log.Err(err).Msg("failed to send error message")
			return
		}
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   response,
	})
	if err != nil {
		log.Err(err).Msg("failed to send response message")
		return
	}
}

func userStatusHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	UserName := update.Message.From.Username
	ChatId := update.Message.Chat.ID

	log.Info().Msgf("User %s with chat id %d set status", UserName, ChatId)

	prompt := strings.TrimSpace(update.Message.Text[len("/status"):])
	prompt = fmt.Sprintf("User %s with chat id %d wrote message: %s", UserName, ChatId, prompt)
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: ChatId,
		Text:   prompt,
	})
	if err != nil {
		log.Err(err).Msg("failed to send message")
		return
	}
}

func echoHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	prompt := strings.TrimSpace(update.Message.Text[len("/echo"):])
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   prompt,
	})
	if err != nil {
		log.Err(err).Msg("failed to send message")
		return
	}
}
