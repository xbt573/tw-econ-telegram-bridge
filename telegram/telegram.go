package telegram

import (
	"context"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

type Telegram struct {
	serverName  string
	chatId      int64
	threadId    int64
	bot         *gotgbot.Bot
	updater     *ext.Updater
	receiveChan chan string
	sendChan    chan string
}

type TelegramOpts struct {
	Token       string
	ServerName  string
	ThreadId    int64
	ChatId      int64
	ReceiveChan chan string
	SendChan    chan string
	BotOpts     *gotgbot.BotOpts
}

func NewTelegram(opts TelegramOpts) (*Telegram, error) {
	bot, err := gotgbot.NewBot(opts.Token, opts.BotOpts)
	if err != nil {
		return nil, err
	}

	telegram := &Telegram{
		bot:         bot,
		chatId:      opts.ChatId,
		serverName:  opts.ServerName,
		threadId:    opts.ThreadId,
		sendChan:    opts.SendChan,
		receiveChan: opts.ReceiveChan,
	}

	telegram.updater = ext.NewUpdater(&ext.UpdaterOpts{
		Dispatcher: ext.NewDispatcher(&ext.DispatcherOpts{
			Error: func(bot *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
				slog.Error(
					"Failed to process update!",
					slog.String(
						"err",
						err.Error(),
					),
				)
				return ext.DispatcherActionNoop
			},
		}),
	})

	telegram.updater.Dispatcher.AddHandler(handlers.NewCommand("currentthreadid", telegram.GetThreadId))
	telegram.updater.Dispatcher.AddHandler(handlers.NewMessage(message.Text, telegram.OnText))
	telegram.updater.Dispatcher.AddHandler(handlers.NewMessage(message.All, telegram.OnMedia))

	return telegram, nil
}

func (t *Telegram) OnText(bot *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveMessage.MessageThreadId != t.threadId {
		return nil
	}

	username := ctx.EffectiveSender.Name()
	text := strings.ReplaceAll(ctx.EffectiveMessage.Text, "\"", "\\\"")
	text = ReplaceFromEmoji(text)

	t.sendChan <- fmt.Sprintf("%v: %v", username, text)
	return nil
}
func (t *Telegram) OnMedia(bot *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveMessage.MessageThreadId != t.threadId {
		return nil
	}

	username := ctx.EffectiveSender.Name()
	text := strings.ReplaceAll(ctx.EffectiveMessage.Caption, "\"", "\\\"")
	text = ReplaceFromEmoji(text)

	t.sendChan <- fmt.Sprintf("%v: [MEDIA] %v", username, text)
	return nil
}

func (t *Telegram) GetThreadId(bot *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(bot, strconv.FormatInt(ctx.EffectiveMessage.MessageThreadId, 10), nil)
	if err != nil {
		return err
	}

	return nil
}

func (t *Telegram) Start(ctx context.Context) error {
	err := t.updater.StartPolling(t.bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})

	if err != nil {
		return err
	}

	go t.updater.Idle()
	defer t.updater.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case x := <-t.receiveChan:
			msg := ReplaceToEmoji(x)
			_, err := t.bot.SendMessage(t.chatId, msg, &gotgbot.SendMessageOpts{
				MessageThreadId: t.threadId,
			})
			if err != nil {
				return err
			}
		}
	}
}
