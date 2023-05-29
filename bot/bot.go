package bot

import (
	"fmt"
	"regexp"
	"strings"
	"tw-econ-telegram-bridge/econ"
	"tw-econ-telegram-bridge/telegram"
)

var chatRegex = regexp.MustCompile(`\[?chat\]?: \d+:-?\d+:(.*)`)

type Bot struct {
	econ *econ.ECON
	tg   *telegram.Telegram

	closed chan bool
}

func NewBot(econ *econ.ECON, tg *telegram.Telegram) *Bot {
	return &Bot{
		econ:   econ,
		tg:     tg,
		closed: make(chan bool),
	}
}

func (b *Bot) Start(errch chan error) {
	msgch := make(chan string)

	b.tg.Subscribe(msgch)
	defer b.tg.Unsubscribe(msgch)

	go func() {
		for b.econ.Connected {
			message, err := b.econ.Read()
			if err != nil {
				errch <- err
				continue
			}
			if strings.Contains(message, "chat") {
				match := chatRegex.FindStringSubmatch(message)
				if len(match) == 0 {
					continue
				}

				err := b.tg.Publish(fmt.Sprintf("[%v]\n%v", b.econ.ServerName, match[1]))
				if err != nil {
					errch <- err
				}
			}
		}
	}()

	go func() {
	outer:
		for {
			msg := <-msgch

			if len(strings.Split(msg, "\n")) == 1 {
				err := b.econ.Send(msg)
				if err != nil {
					errch <- err
					continue
				}
				continue
			}

			name, text, found := strings.Cut(msg, ": ")
			if !found {
				continue
			}

			err := b.econ.Send(fmt.Sprintf("%v:", name))
			if err != nil {
				errch <- err
				continue
			}

			lines := strings.Split(text, "\n")
			for _, line := range lines {
				err := b.econ.Send(">         " + line)

				if err != nil {
					errch <- err
					continue outer
				}
			}
		}
	}()

	<-b.closed
}

func (b *Bot) Close() {
	b.closed <- true
	b.econ.Disconnect()
}
