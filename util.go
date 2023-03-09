package main

import (
	"fmt"
	"os"

	"golang.org/x/exp/slog"
)

func getEnv(name string) string {
	env, exists := os.LookupEnv(name)
	if !exists {
		slog.Error(fmt.Sprintf("%v not set", name), nil)

	}

	return env
}

func getEnvDefault(name string, defaultValue string) string {
	env, exists := os.LookupEnv(name)
	if !exists {
		return defaultValue
	}

	return env
}
