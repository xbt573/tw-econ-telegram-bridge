package telegram

import (
	"fmt"
	"strings"

	"gopkg.in/telebot.v3"
)

type Telegram struct {
	client *telebot.Bot
	chatId int64

	listeners []chan string // econs who are listening for tg
}

func NewTelegram(settings telebot.Settings, chatId int64) (*Telegram, error) {
	bot, err := telebot.NewBot(settings)
	if err != nil {
		return nil, err
	}

	return &Telegram{
		client:    bot,
		listeners: make([]chan string, 0),
		chatId:    chatId,
	}, nil
}

func (t *Telegram) Listen() {
	t.client.Handle(telebot.OnText, func(ctx telebot.Context) error {
		if ctx.Chat().ID != t.chatId {
			return nil
		}

		username := strings.Join(
			[]string{
				ctx.Message().Sender.FirstName,
				ctx.Message().Sender.LastName,
			},
			" ",
		)

		text := ReplaceFromEmoji(ctx.Message().Text)
		text = strings.ReplaceAll(text, "\"", "\\\"")

		t.broadcast(fmt.Sprintf("%v: %v", username, text))
		return nil
	})

	t.client.Handle(telebot.OnMedia, func(ctx telebot.Context) error {
		username := strings.Join(
			[]string{
				ctx.Message().Sender.FirstName,
				ctx.Message().Sender.LastName,
			},
			" ",
		)

		attachmentType := ""
		switch {
		case ctx.Message().Animation != nil:
			attachmentType = "ANIMATION"

		case ctx.Message().Audio != nil:
			attachmentType = "AUDIO"

		case ctx.Message().Photo != nil:
			attachmentType = "PHOTO"

		case ctx.Message().Sticker != nil:
			attachmentType = "STICKER"

		case ctx.Message().Video != nil:
			attachmentType = "VIDEO"

		case ctx.Message().VideoNote != nil:
			attachmentType = "VIDEO NOTE"

		case ctx.Message().Voice != nil:
			attachmentType = "VOICE"
		}

		text := ReplaceFromEmoji(ctx.Message().Text)
		text = strings.ReplaceAll(text, "\"", "\\\"")

		t.broadcast(
			fmt.Sprintf(
				"%v: [%v] %v",
				username,
				attachmentType,
				text,
			),
		)
		return nil
	})

	t.client.Start()
}

func (t *Telegram) Subscribe(ch chan string) {
	t.listeners = append(t.listeners, ch)
}

func (t *Telegram) Unsubscribe(ch chan string) {
	for i, v := range t.listeners {
		if v == ch {
			t.listeners = append(t.listeners[:i], t.listeners[i+1:]...)
			return
		}
	}
}

func (t *Telegram) broadcast(msg string) {
	for _, listener := range t.listeners {
		listener <- msg
	}
}

func (t *Telegram) Publish(msg string) error {
	_, err := t.client.Send(RecipientFromInt64(t.chatId), ReplaceToEmoji(msg))
	if err != nil {
		return err
	}

	return nil
}
