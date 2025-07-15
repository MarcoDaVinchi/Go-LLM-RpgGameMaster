package main

import (
	"context"
	"os"
	"os/signal"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"

	"go-llm-rpggamemaster/llm" // Замените на ваш актуальный module path при необходимости
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(echoHandler),
	}
	var token = os.Getenv("RPG_TELEGRAM_BOT_API_KEY")
	if token == "" {
		log.Fatal().Msg("RPG_TELEGRAM_BOT_API_KEY is not set")
	}
	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/info", bot.MatchTypePrefix, infoHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/web", bot.MatchTypePrefix, webHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/gpt", bot.MatchTypePrefix, gptHandler)          // Исправленный ключ
	b.RegisterHandler(bot.HandlerTypeMessageText, "/local", bot.MatchTypePrefix, localModelHandler) // Новый обработчик

	b.Start(ctx)
}

// Новый хендлер для сообщений с /local
func localModelHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	prompt := strings.TrimSpace(update.Message.Text[len("/local"):]) // Берет текст после ключа
	client := llm.NewLocalModelClient("")                            // Использует дефолтную модель
	response, err := client.GenerateSimpleResponse(ctx, prompt)
	if err != nil {
		log.Err(err).Msg("failed to get response from local model")
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ошибка обращения к локальной модели: " + err.Error(),
		})
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   response,
	})
}

// Новый хендлер для сообщений с /gpt
func gptHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	prompt := strings.TrimSpace(update.Message.Text[len("/gpt"):]) // Берет текст после ключа
	client := llm.NewOpenAIClient("", "")                          // Использует OPENAI_API_KEY и дефолтную модель
	response, err := client.GenerateSimpleResponse(ctx, prompt)
	if err != nil {
		log.Err(err).Msg("failed to get response from OpenAI")
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ошибка обращения к OpenAI: " + err.Error(),
		})
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   response,
	})
}

// infoHandler - выводит список администраторов чата
func infoHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	me, getMeErr := b.GetMe(ctx)
	if getMeErr != nil {
		log.Err(getMeErr).Msg("failed to get me")
		return
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   me.FirstName,
	})
	if err != nil {
		log.Err(err).Msg("failed to send message in infoHandler")
	}
}

// webHandler - пример обработки команды /web
func webHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	text := update.Message.Text
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Вы вызвали команду /web! Текст сообщения: " + text,
	})
	if err != nil {
		log.Err(err).Msg("failed to send message in webHandler")
	}
}

// echoHandler - отвечает тем же текстом
func echoHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
	if err != nil {
		log.Err(err).Msg("failed to send message")
		return
	}
}
