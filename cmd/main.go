package main

import (
	"context"
	botapp "my-go-project/bot"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type RealBot struct {
    bot *bot.Bot
}

func (r *RealBot) SendMessage(ctx context.Context, chatID int64, text string) {
    r.bot.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: chatID,
        Text:   text,
    })
}

func main() {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        panic("TELEGRAM_BOT_TOKEN not set")
    }

    b, err := bot.New(token)
    if err != nil {
        panic(err)
    }

    conv := botapp.NewConversation()
    wrapper := &RealBot{bot: b}

    b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact,
        func(ctx context.Context, b *bot.Bot, update *models.Update) {
            conv.HandleStart(ctx, wrapper, update.Message.Chat.ID)
        })

    b.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypePrefix,
        func(ctx context.Context, b *bot.Bot, update *models.Update) {
            conv.HandleMessage(ctx, wrapper, update)
        })

    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()

    b.Start(ctx)
}
