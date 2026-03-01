package main

import (
	"bot/client/aiclient"
	"bot/client/serverclient"
	"bot/internal/route"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	token := os.Getenv("TELEGRAM_TOKEN")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	AIClient, err := aiclient.NewClient(os.Getenv("KAFKA_BROKERS"))
	if err != nil {
		log.Println("error to connect AI microservis")
	}

	grpcClient, err := serverclient.NewClient(os.Getenv("SERVER_HOST"))
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		go route.HandleMessage(grpcClient, bot, update.Message, AIClient)
	}
}
