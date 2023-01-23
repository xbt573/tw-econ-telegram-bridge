package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"
	"tw-econ-telegram-bridge/econ"

	"gopkg.in/telebot.v3"
)

var (
	chatRegex  = regexp.MustCompile(`\[chat\]: \d+:-?\d+:(.*)`)
	host       = getEnvDefault("TW_HOST", "localhost")
	serverName = getEnv("SERVER_NAME")
	port       = intMustParse(getEnvDefault("TW_PORT", "8303"))
	password   = getEnv("TW_PASSWORD")
	token      = getEnv("API_TOKEN")
	chatId     = getEnv("CHAT_ID")
	console    = econ.NewECON(host, password, port)
)

func init() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
}

func main() {
	log.Println("Starting tw-econ-telegram-bridge...")

	bot, err := telebot.NewBot(telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatalln(err)
	}

	bot.Handle(telebot.OnText, func(ctx telebot.Context) error {
		username := strings.Join(
			[]string{
				ctx.Message().Sender.FirstName,
				ctx.Message().Sender.LastName,
			},
			" ",
		)

		message := fmt.Sprintf("%v: %v", username, ctx.Message().Text)
		err := console.Send(message)
		if err != nil {
			return err
		}

		return nil
	})

	err = console.Connect()
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		for console.Connected {
			message, err := console.Read()
			if err != nil {
				log.Fatalln(err)
			}
			if strings.Contains(message, "chat") {
				match := chatRegex.FindStringSubmatch(message)
				if len(match) == 0 {
					continue
				}

				message := fmt.Sprintf("[%v] %v", serverName, match[1])
				_, err := bot.Send(FakeRecipient{ID: chatId}, message)
				if err != nil {
					log.Fatalln(err)
				}
			}
		}
	}()

	go bot.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
	case <-console.Completed:
		break

	case <-c:
		break
	}

	log.Println("Shutting down...")
}
