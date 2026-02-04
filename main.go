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
	"go-llm-rpggamemaster/retrievers"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var llmProvider interfaces.InferenceProvider
var retriever retrievers.Retriever

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go func() {
		<-ctx.Done()
		log.Info().Msg("Shutting down...")
		if retriever != nil {
			if closer, ok := retriever.(interface{ Close() }); ok {
				closer.Close()
			}
		}
	}()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	profile := cfg.Profile
	if profile == "local" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	providerFactory := factory.NewProviderFactory(cfg)
	llmProvider, err = providerFactory.CreateInferenceProvider()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create LLM provider")
	}

	log.Info().Msgf("Using LLM provider: %s", llmProvider.Name())

	if cfg.VectorRetriever.Type != 0 {
		providerFactory := factory.NewProviderFactory(cfg)
		embedder, err := providerFactory.CreateEmbeddingProvider()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create embedding provider")
		}

		log.Info().Msgf("Using embedding provider: %s", embedder.Name())

		retriever, err = providerFactory.CreateRetriever(embedder, strings.ToLower(cfg.VectorRetriever.Type.String()))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create retriever")
		}

		log.Info().Msgf("Retriever initialized: %s", cfg.VectorRetriever.Type)
	}

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

	messages := []interfaces.Message{
		{Role: "user", Content: prompt},
	}
	response, err := llmProvider.GenerateResponse(ctx, messages, 0.7, 0)
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
