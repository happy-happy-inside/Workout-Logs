package route

import (
	"bot/client"
	action "bot/internal/botaction"
	hand "bot/internal/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMessage(grpcClient *client.Client, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	switch msg.Command() {

	case "start":
		hand.HandleStart(bot, msg)

	case "help":
		hand.HandleHelp(bot, msg)

	case "add":
		hand.HandleAdd(grpcClient, bot, msg)

	case "get":
		hand.HandleGet(grpcClient, bot, msg)

	case "top":
		hand.HandleTop(grpcClient, bot, msg)

	default:
		action.Send(bot, msg.Chat.ID, "Неизвестная команда. Используй /help")
	}
}
