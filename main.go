package main

import (
	"context"
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

	b.Start(ctx)
}

func gptHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	prompt := strings.TrimSpace(update.Message.Text[len("/gpt"):])
	if prompt == "" {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Пожалуйста, укажите запрос после команды /gpt",
		})
		return
	}

	// Создаем клиент и настраиваем его (только конфигурация)
	client := llm.NewRPGOllamaClient("llama3.1")

	response, err := client.GenerateSimpleResponse(ctx, prompt)
	if err != nil {
		log.Err(err).Msg("failed to get response from Ollama")
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ошибка обращения к Ollama: " + err.Error(),
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   response,
	})
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
