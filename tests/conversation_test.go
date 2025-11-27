package tests

import (
	"context"
	botapp "my-go-project/bot"
	"testing"

	"github.com/go-telegram/bot/models"
)

/* -------------------------------------------------------------------------- */
/*                                  MOCK BOT                                   */
/* -------------------------------------------------------------------------- */

type MockBot struct {
    Sent []SentMessage
}

type SentMessage struct {
    ChatID int64
    Text   string
}

func (m *MockBot) SendMessage(ctx context.Context, chatID int64, text string) {
    m.Sent = append(m.Sent, SentMessage{ChatID: chatID, Text: text})
}

func newUpdate(text string) *models.Update {
    return &models.Update{
        Message: &models.Message{
            Chat: models.Chat{ID: 1},
            Text: text,
        },
    }
}

var ctx = context.Background()

/* -------------------------------------------------------------------------- */
/*                                    TESTS                                   */
/* -------------------------------------------------------------------------- */

func TestStartDialog(t *testing.T) {
    conv := botapp.NewConversation()
    bot := &MockBot{}

    conv.HandleStart(ctx, bot, 1)

    if len(bot.Sent) == 0 {
        t.Fatal("expected greeting message, got nothing")
    }

    if bot.Sent[0].Text == "" {
        t.Fatal("greeting message is empty")
    }
}

func TestSaveAge(t *testing.T) {
    conv := botapp.NewConversation()
    bot := &MockBot{}

    conv.HandleMessage(ctx, bot, newUpdate("Age"))
    conv.HandleMessage(ctx, bot, newUpdate("25"))

    data := conv.UserData.Get(1)
    if data["age"] != "25" {
        t.Fatalf("expected age=25, got %#v", data)
    }
}

func TestSaveColour(t *testing.T) {
    conv := botapp.NewConversation()
    bot := &MockBot{}

    conv.HandleMessage(ctx, bot, newUpdate("Favourite colour"))
    conv.HandleMessage(ctx, bot, newUpdate("green"))

    data := conv.UserData.Get(1)
    if data["favourite colour"] != "green" {
        t.Fatalf("expected colour=green, got %#v", data)
    }
}

func TestSaveSiblings(t *testing.T) {
    conv := botapp.NewConversation()
    bot := &MockBot{}

    conv.HandleMessage(ctx, bot, newUpdate("Number of siblings"))
    conv.HandleMessage(ctx, bot, newUpdate("3"))

    data := conv.UserData.Get(1)
    if data["number of siblings"] != "3" {
        t.Fatalf("expected siblings=3, got %#v", data)
    }
}

func TestSaveCustomCategory(t *testing.T) {
    conv := botapp.NewConversation()
    bot := &MockBot{}

    conv.HandleMessage(ctx, bot, newUpdate("Something else..."))
    conv.HandleMessage(ctx, bot, newUpdate("Hobby"))
    conv.HandleMessage(ctx, bot, newUpdate("chess"))

    data := conv.UserData.Get(1)
    if data["hobby"] != "chess" {
        t.Fatalf("expected hobby=chess, got %#v", data)
    }
}

func TestChangeExistingValue(t *testing.T) {
    conv := botapp.NewConversation()
    bot := &MockBot{}

    // первая запись
    conv.HandleMessage(ctx, bot, newUpdate("Age"))
    conv.HandleMessage(ctx, bot, newUpdate("20"))

    // изменение
    conv.HandleMessage(ctx, bot, newUpdate("Age"))
    conv.HandleMessage(ctx, bot, newUpdate("32"))

    data := conv.UserData.Get(1)
    if data["age"] != "32" {
        t.Fatalf("expected changed age=32, got %#v", data)
    }
}

func TestPersistenceAfterNewStart(t *testing.T) {
    conv := botapp.NewConversation()
    bot := &MockBot{}

    // сохраняем данные
    conv.HandleMessage(ctx, bot, newUpdate("Age"))
    conv.HandleMessage(ctx, bot, newUpdate("44"))

    // новый /start НЕ должен стирать данные
    conv.HandleStart(ctx, bot, 1)

    data := conv.UserData.Get(1)
    if data["age"] != "44" {
        t.Fatalf("expected age to persist after /start, got %#v", data)
    }
}
