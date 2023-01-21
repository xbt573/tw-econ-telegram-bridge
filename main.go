package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"time"

	"gopkg.in/telebot.v3"
)

var (
	chatRegex = regexp.MustCompile(`\[chat\]: \d+:-?\d+:(.*)`)
)

type FakeRecipient struct {
	ID string
}

func (f FakeRecipient) Recipient() string {
	return f.ID
}

func main() {
	host, exists := os.LookupEnv("TW_HOST")
	if !exists {
		log.Fatalln("TW_HOST not set")
	}

	hostName, exists := os.LookupEnv("SERVER_NAME")
	if !exists {
		log.Fatalln("SERVER_NAME not set")
	}

	port, exists := os.LookupEnv("TW_PORT")
	if !exists {
		log.Fatalln("TW_PORT not set")
	}

	password, exists := os.LookupEnv("TW_PASSWORD")
	if !exists {
		log.Fatalln("TW_PASSWORD not set")
	}

	token, exists := os.LookupEnv("API_TOKEN")
	if !exists {
		log.Fatalln("API_TOKEN not set")
	}

	chatID, exists := os.LookupEnv("CHAT_ID")
	if !exists {
		log.Fatalln("CHAT_ID not set")
	}

	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		log.Fatalln(err)
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Fatalln(err)
	}

	message := string(buffer[:n])
	if strings.Contains(message, "Enter password") {
		_, err = conn.Write([]byte(password + "\n"))
		if err != nil {
			log.Fatalln(err)
		}

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			log.Fatalln(err)
		}

		message := string(buffer[:n])
		if !strings.Contains(message, "Authentication successful") {
			log.Fatalln("Wrong password or timeout")
		}
	}

	bot, err := telebot.NewBot(telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	mutex := sync.Mutex{}

	bot.Handle(telebot.OnText, func(ctx telebot.Context) error {
		username := strings.Join(
			[]string{
				ctx.Message().Sender.FirstName,
				ctx.Message().Sender.LastName,
			},
			" ",
		)

		message := fmt.Sprintf("%v: %v", username, ctx.Message().Text)

		mutex.Lock()
		defer mutex.Unlock()
		// _, err := conn.Write([]byte(`say "` + message + `"\n`))
		_, err := conn.Write([]byte(`say "` + message + `"` + "\n"))
		if err != nil {
			return err
		}

		return nil
	})

	go func() {
		for conn != nil {
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				log.Fatalln(err)
			}

			message = string(buffer[:n])
			if strings.Contains(message, "chat") {
				match := chatRegex.FindStringSubmatch(message)
				if len(match) == 0 {
					continue
				}

				message := fmt.Sprintf("[%v] %v", hostName, match[1])
				_, err := bot.Send(FakeRecipient{ID: chatID}, message)
				if err != nil {
					log.Fatalln(err)
				}
			}
		}
	}()

	go bot.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		conn.Close()
		conn = nil

		bot.Stop()
		break
	}
	log.Println("Shutting down...")
}
