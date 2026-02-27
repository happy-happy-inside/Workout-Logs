package botaction

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func Send(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	message := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(message)
	return err
}

func SendSticker(bot *tgbotapi.BotAPI, chatID int64, fileID string) error {
	sticker := tgbotapi.NewSticker(chatID, tgbotapi.FileID(fileID))
	_, err := bot.Send(sticker)
	return err
}
