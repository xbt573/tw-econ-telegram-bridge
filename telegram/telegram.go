package telegram

import (
	"fmt"
	"strings"

	"gopkg.in/telebot.v3"
)

type Telegram struct {
	client *telebot.Bot
	chatId string

	listeners []chan string // econs who are listening for tg
}

func NewTelegram(settings telebot.Settings, chatId string) (*Telegram, error) {
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
		username := strings.Join(
			[]string{
				ctx.Message().Sender.FirstName,
				ctx.Message().Sender.LastName,
			},
			" ",
		)

		t.broadcast(fmt.Sprintf("%v: %v", username, ctx.Message().Text))
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
	_, err := t.client.Send(FakeRecipient{ID: t.chatId}, msg, &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdownV2,
	})
	if err != nil {
		return err
	}

	return nil
}
