package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"go-llm-rpggamemaster/llm" // Импортируем наш пакет llm

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(gptHandler),
	}
	var token = os.Getenv("RPG_TELEGRAM_BOT_API_KEY")
	if token == "" {
		log.Fatal().Msg("RPG_TELEGRAM_BOT_API_KEY is not set")
	}
	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}

	// Регистрируем обработчик для команды /gpt
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

	// Создаем клиент и настраиваем его (только конфигурация)
	client := llm.NewRPGOllamaClient("llama3.1")

	response, err := client.GenerateSimpleResponse(ctx, prompt)
	if err != nil {
		log.Err(err).Msg("failed to get response from Ollama")
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ошибка обращения к Ollama: " + err.Error(),
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
