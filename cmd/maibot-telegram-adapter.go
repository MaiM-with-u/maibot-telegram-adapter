package main

import (
	"log"

	"github.com/MaiM-with-u/maibot-telegram-adapter/internal/config"
	"github.com/MaiM-with-u/maibot-telegram-adapter/internal/maibot"
	"github.com/MaiM-with-u/maibot-telegram-adapter/internal/telegram"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	config.Load()

	err := maibot.InitDefaultClient()
	if err != nil {
		log.Printf("Warning: Failed to initialize MaiBot client: %v", err)
	}

	// Set up message handler for incoming messages from MaiBot server
	client := maibot.GetDefaultClient()
	if client != nil {
		client.SetMessageHandler(func(messageBase *maibot.MessageBase) {
			// Forward MessageBase messages from MaiBot server to Telegram
			err := telegram.SendMessageToTelegram(messageBase)
			if err != nil {
				log.Printf("Failed to forward message to Telegram: %v", err)
			}
		})
	}

	go telegram.StartBot()

	select {}
}
