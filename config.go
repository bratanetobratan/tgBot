package main

import (
	"os"
)

func GetBotToken() string {
	token := os.Getenv("tgBotToken")
	return token
}
