package main

import "github.com/xbt573/tw-econ-telegram-bridge/econ"

type Config struct {
	ChatId   int64           `yaml:"chat_id"`
	ThreadId int64           `yaml:"thread_id"`
	Ip       string          `yaml:"ip"`
	Port     uint16          `yaml:"port"`
	Password string          `yaml:"password"`
	Token    string          `yaml:"token"`
	Type     econ.ServerType `yaml:"type"`
}
