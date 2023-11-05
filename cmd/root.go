package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xbt573/tw-econ-telegram-bridge/bot"
	"github.com/xbt573/tw-econ-telegram-bridge/econ"
	"github.com/xbt573/tw-econ-telegram-bridge/telegram"
	"log/slog"
	"os"
	"os/signal"
)

var (
	cfgFile string
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(
		&cfgFile,
		"config",
		"c",
		"",
		"Config file for bridge",
	)

	rootCmd.PersistentFlags().Int64("chatid", 0, "Forum chat ID")
	rootCmd.PersistentFlags().Int64("threadid", 0, "Target thread ID")
	rootCmd.PersistentFlags().String("ip", "", "Server IP address")
	rootCmd.PersistentFlags().Uint16("port", 0, "Server port")
	rootCmd.PersistentFlags().String("password", "", "Server ECON password")
	rootCmd.PersistentFlags().String("token", "", "Telegram bot API token")
	rootCmd.PersistentFlags().String(
		"type",
		"",
		"Server type (one of 'ddnet', 'trainfng', or 'teeworlds'",
	)
	viper.BindPFlag("chat_id", rootCmd.PersistentFlags().Lookup("chatid"))
	viper.BindPFlag("thread_id", rootCmd.PersistentFlags().Lookup("threadid"))
	viper.BindPFlag("ip", rootCmd.PersistentFlags().Lookup("ip"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("type", rootCmd.PersistentFlags().Lookup("type"))

	rootCmd.MarkFlagsRequiredTogether(
		"chatid",
		"threadid",
		"ip",
		"port",
		"password",
		"token",
		"type",
	)
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/tw-econ-telegram-bridge")
	viper.AddConfigPath("$XDG_CONFIG_HOME/tw-econ-telegram-bridge")
	viper.AddConfigPath("$HOME/.config/tw-econ-telegram-bridge")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "tw-econ-telegram-bridge",
	Short: "Telegram <-> DDNet (and others) bridge",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Starting bridge...")

		econInstance, err := econ.NewECON(econ.ECONOpts{
			Ip:       viper.GetString("ip"),
			Port:     viper.GetUint16("port"),
			Password: viper.GetString("password"),
		})
		if err != nil {
			slog.Error(
				"Failed to init ECON!",
				slog.String(
					"err",
					err.Error(),
				),
			)
			os.Exit(1)
		}

		var (
			sendChan    = make(chan string)
			receiveChan = make(chan string)
		)

		tgInstance, err := telegram.NewTelegram(telegram.TelegramOpts{
			Token:       viper.GetString("token"),
			ServerName:  viper.GetString("server_name"),
			ThreadId:    viper.GetInt64("thread_id"),
			ChatId:      viper.GetInt64("chat_id"),
			ReceiveChan: receiveChan,
			SendChan:    sendChan,
			BotOpts:     nil,
		})
		if err != nil {
			slog.Error(
				"Failed to init Telegram!",
				slog.String(
					"err",
					err.Error(),
				),
			)
			os.Exit(1)
		}

		botInstance := bot.NewBot(bot.BotOpts{
			Econ:        econInstance,
			ServerType:  econ.ServerType(viper.GetString("type")),
			ReceiveChan: sendChan,
			SendChan:    receiveChan,
		})
		if err != nil {
			slog.Error(
				"Failed to init bot!",
				slog.String(
					"err",
					err.Error(),
				),
			)
			os.Exit(1)
		}

		errch := make(chan error)

		sigch := make(chan os.Signal)
		signal.Notify(sigch, os.Interrupt)

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			err := tgInstance.Start(ctx)
			if err != nil {
				errch <- err
			}
		}()

		go func() {
			err := botInstance.Start(ctx)
			if err != nil {
				errch <- err
			}
		}()

		slog.Info("Started!")

		select {
		case <-sigch:
			slog.Info("Caught interrupt, shutting down...")
			cancel()
		case x := <-errch:
			slog.Error(
				"Caught error!",
				slog.String(
					"err",
					x.Error(),
				),
			)
			cancel()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
