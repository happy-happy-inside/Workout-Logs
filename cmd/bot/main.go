package main

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Massage struct {
}

func Connect() error {

}

func ParsMassage() *Massage {

}

func ChoiseFunc() {

}

func main() {
	bot, err := tgbotapi.NewBotAPI("TOKEN")
	if err != nil {
		log.Fatal(err)
	}

	proto.AddRes

	updates := bot.GetUpdatesChan(tgbotapi.UpdateConfig{Timeout: 60})

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// вызываем backend
		resp, err := client.Hello(
			context.Background(),
			&pb.HelloRequest{
				Name: update.Message.From.UserName,
			},
		)

		if err != nil {
			log.Println(err)
			continue
		}

		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			resp.Message,
		)

		bot.Send(msg)
	}
}
