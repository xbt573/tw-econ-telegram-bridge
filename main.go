package main

import (
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"
	"tw-econ-telegram-bridge/bot"
	"tw-econ-telegram-bridge/econ"
	"tw-econ-telegram-bridge/telegram"

	"golang.org/x/exp/slog"
	"gopkg.in/telebot.v3"
)

var (
	portsRegex = regexp.MustCompile(`(\d+:\d+|\d+),?`)
	rangeRegex = regexp.MustCompile(`(\d+):(\d+)`)
	host       = getEnvDefault("TW_HOST", "localhost")
	// serverName = getEnv("SERVER_NAME")
	ports    = getEnvDefault("TW_PORT", "8303")
	password = getEnv("TW_PASSWORD")
	token    = getEnv("API_TOKEN")
	chatId   = getEnv("CHAT_ID")
)

func main() {
	slog.Info("Starting service...")

	tg, err := telegram.NewTelegram(telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: time.Second * 5},
	}, chatId)
	if err != nil {
		slog.Error("Failed creating bot", err)
		os.Exit(1)
	}

	slog.Info("Initialized bot...")

	match := portsRegex.FindAllStringSubmatch(ports, -1)
	if len(match) == 0 {
		slog.Error("Incorrect ports definition, check documentation", nil)
		os.Exit(1)
	}

	intPorts := []int{}

	for _, portMatch := range match {
		if strings.Contains(portMatch[1], ":") {
			rangeMatch := rangeRegex.FindAllStringSubmatch(portMatch[1], -1)
			if len(rangeMatch) == 0 {
				continue
			}

			first, err := strconv.Atoi(rangeMatch[0][1])
			if err != nil {
				continue
			}

			second, err := strconv.Atoi(rangeMatch[0][2])
			if err != nil {
				continue
			}

			for i := first; i <= second; i++ {
				intPorts = append(intPorts, i)
			}

			continue
		}

		port, err := strconv.Atoi(portMatch[1])
		if err != nil {
			continue
		}

		intPorts = append(intPorts, port)
	}

	if len(intPorts) == 0 {
		slog.Error("No ports parsed. Check documentation for format", nil)
		os.Exit(1)
	}

	errch := make(chan error)
	workers := []*bot.Bot{}

	for _, port := range intPorts {
		econ := econ.NewECON(host, password, port)
		err := econ.Connect()
		if err != nil {
			slog.Error(
				"Failed connecting to ECON",
				err,
				slog.String("host", host),
				slog.Int("port", port),
			)
			os.Exit(1)
		}

		worker := bot.NewBot(econ, tg)

		go worker.Start(errch)
		workers = append(workers, worker)
	}

	go tg.Listen()

	go func() {
		for {
			err := <-errch

			slog.Error("error occured", err)
		}
	}()

	slog.Info("Started!")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	slog.Info("Shutting down...")

	for _, worker := range workers {
		worker.Close()
	}
}
