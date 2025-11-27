package botapp

import (
	"context"
	"strings"
	"sync"

	"github.com/go-telegram/bot/models"
)

/* -------------------------------------------------------------------------- */
/*                              UserData Storage                               */
/* -------------------------------------------------------------------------- */

type UserData struct {
    mu   sync.Mutex
    data map[int64]map[string]string
}

func NewUserData() *UserData {
    return &UserData{
        data: make(map[int64]map[string]string),
    }
}

func (u *UserData) Get(chatID int64) map[string]string {
    u.mu.Lock()
    defer u.mu.Unlock()

    m, ok := u.data[chatID]
    if !ok {
        m = make(map[string]string)
        u.data[chatID] = m
    }
    return m
}

func FactsToStr(m map[string]string) string {
    if len(m) == 0 {
        return ""
    }
    parts := []string{}
    for k, v := range m {
        parts = append(parts, k+" - "+v)
    }
    return "\n" + strings.Join(parts, "\n") + "\n"
}

/* -------------------------------------------------------------------------- */
/*                                  Keyboard                                   */
/* -------------------------------------------------------------------------- */

var ReplyKeyboard = [][]string{
    {"Age", "Favourite colour"},
    {"Number of siblings", "Something else..."},
    {"Done"},
}

func ContainsKeyboard(s string) bool {
    for _, row := range ReplyKeyboard {
        for _, v := range row {
            if v == s {
                return true
            }
        }
    }
    return false
}

/* -------------------------------------------------------------------------- */
/*                                Bot Interface                                */
/* -------------------------------------------------------------------------- */

type BotAPI interface {
    SendMessage(ctx context.Context, chatID int64, text string)
}

/* -------------------------------------------------------------------------- */
/*                                Conversation                                 */
/* -------------------------------------------------------------------------- */

type Conversation struct {
    UserData  *UserData
    ChoiceKey map[int64]string
    muChoice  sync.Mutex
}

func NewConversation() *Conversation {
    return &Conversation{
        UserData:  NewUserData(),
        ChoiceKey: make(map[int64]string),
    }
}

func (c *Conversation) getChoice(chatID int64) string {
    c.muChoice.Lock()
    defer c.muChoice.Unlock()
    return c.ChoiceKey[chatID]
}

func (c *Conversation) setChoice(chatID int64, v string) {
    c.muChoice.Lock()
    c.ChoiceKey[chatID] = v
    c.muChoice.Unlock()
}

func (c *Conversation) clearChoice(chatID int64) {
    c.muChoice.Lock()
    delete(c.ChoiceKey, chatID)
    c.muChoice.Unlock()
}

/* -------------------------------------------------------------------------- */
/*                              Handlers (Testable)                             */
/* -------------------------------------------------------------------------- */

func (c *Conversation) HandleStart(ctx context.Context, b BotAPI, chatID int64) {
    data := c.UserData.Get(chatID)

    reply := "Hi! My name is Doctor Botter."
    if len(data) > 0 {
        keys := []string{}
        for k := range data {
            keys = append(keys, k)
        }
        reply += " You already told me your " + strings.Join(keys, ", ") +
            ". Why don't you tell me something more about yourself? Or change anything I already know."
    } else {
        reply += " I will hold a more complex conversation with you. Why don't you tell me something about yourself?"
    }

    b.SendMessage(ctx, chatID, reply)
}

func (c *Conversation) HandleMessage(ctx context.Context, b BotAPI, upd *models.Update) {
    chatID := upd.Message.Chat.ID
    text := upd.Message.Text
    lower := strings.ToLower(text)

    data := c.UserData.Get(chatID)
    prevChoice := c.getChoice(chatID)

    switch {
    case text == "Done":
        b.SendMessage(ctx, chatID, "I learned these facts about you:"+FactsToStr(data)+"Until next time!")
        c.clearChoice(chatID)

    case ContainsKeyboard(text) && text != "Something else...":
        c.setChoice(chatID, lower)

        if val, ok := data[lower]; ok {
            b.SendMessage(ctx, chatID, "Your "+lower+"? I already know the following about that: "+val)
        } else {
            b.SendMessage(ctx, chatID, "Your "+lower+"? Yes, I would love to hear about that!")
        }

    case text == "Something else...":
        c.setChoice(chatID, "")
        b.SendMessage(ctx, chatID, "Alright, please send me the category first.")

    default:
        if prevChoice == "" {
            c.setChoice(chatID, lower)
            b.SendMessage(ctx, chatID, "Great — now send me the info for that category.")
        } else {
            data[prevChoice] = lower
            c.clearChoice(chatID)

            msg := "Neat! Just so you know, this is what you already told me:" +
                FactsToStr(data) +
                "You can tell me more, or change your opinion on something."

            b.SendMessage(ctx, chatID, msg)
        }
    }
}
