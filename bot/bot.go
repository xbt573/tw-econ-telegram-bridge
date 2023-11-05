package bot

import (
	"context"
	"github.com/xbt573/tw-econ-telegram-bridge/econ"
)

type Bot struct {
	econ        *econ.ECON
	serverType  econ.ServerType
	receiveChan chan string
	sendChan    chan string
}

type BotOpts struct {
	Econ        *econ.ECON
	ServerType  econ.ServerType
	ReceiveChan chan string
	SendChan    chan string
}

func NewBot(opts BotOpts) *Bot {
	return &Bot{
		econ:        opts.Econ,
		receiveChan: opts.ReceiveChan,
		sendChan:    opts.SendChan,
		serverType:  opts.ServerType,
	}
}

func (b *Bot) Start(ctx context.Context) error {
	err := b.econ.Connect()
	if err != nil {
		return err
	}

	defer b.econ.Disconnect()

	errch := make(chan error)

	go func() {
		for b.econ.Connected() {
			select {
			case <-ctx.Done():
				return
			case x := <-b.receiveChan:
				err := b.econ.Message(x)
				if err != nil {
					errch <- err
					return
				}
			}
		}
	}()

	for b.econ.Connected() {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errch:
			return err
		default:
			msg, err := b.econ.Read()
			if err != nil {
				return err
			}

			text, ok := econ.Adapters[b.serverType].Match(msg)
			if !ok {
				continue
			}

			b.sendChan <- text
		}
	}

	return nil
}
