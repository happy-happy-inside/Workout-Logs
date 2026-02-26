package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	switch msg.Command() {

	case "start":
		handleStart(bot, msg)

	case "help":
		handleHelp(bot, msg)

	case "add":
		handleAdd(bot, msg)

	case "get":
		handleGet(bot, msg)

	case "top":
		handleTop(bot, msg)

	default:
		send(bot, msg.Chat.ID, "Unknown command. Use /help")
	}
}

func send(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func handleStart(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := `<--| Worklout logs |-->
	это сервсис который позволит отслеживать свои тренировочные результаты
	что бы ознакомитсья с функционалом введи команду /help`

	send(bot, msg.Chat.ID, text)
}

func handleHelp(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := `Список команд:
	/start - Приветсвенное сообщение
	/help - Этот список
	/add - Добавить новый результат
	/get - Вывесити нужный результат
	/top - Вывести топ юзеров
	`

	send(bot, msg.Chat.ID, text)
}

type WorkoutInput struct {
	Name   string
	Reps   int
	Sets   int
	Weight float64
	Date   time.Time
}

func handleAdd(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	args := msg.CommandArguments()
	if args == "" {
		send(bot, msg.Chat.ID, "Usage: /add exercise reps sets weight date")
		return
	}

	entries := strings.Split(args, ",")
	var workouts []WorkoutInput

	for _, entry := range entries {
		fields := strings.Fields(strings.TrimSpace(entry))
		if len(fields) != 5 {
			send(bot, msg.Chat.ID, "Invalid format in one of entries")
			return
		}

		reps, err1 := strconv.Atoi(fields[1])
		sets, err2 := strconv.Atoi(fields[2])
		weight, err3 := strconv.ParseFloat(fields[3], 64)
		date, err4 := time.Parse("2006-01-02", fields[4])

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			send(bot, msg.Chat.ID, "Invalid numbers or date format (YYYY-MM-DD)")
			return
		}

		workouts = append(workouts, WorkoutInput{
			Name:   fields[0],
			Reps:   reps,
			Sets:   sets,
			Weight: weight,
			Date:   date,
		})
	}

	send(bot, msg.Chat.ID, fmt.Sprintf("Added %d workout(s)", len(workouts)))
}

func handleGet(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	name := strings.TrimSpace(msg.CommandArguments())
	if name == "" {
		send(bot, msg.Chat.ID, "Usage: /get exercise")
		return
	}

	// Здесь будет gRPC вызов

	send(bot, msg.Chat.ID, "Workout history for "+name)
}

func handleTop(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	arg := strings.TrimSpace(msg.CommandArguments())
	if arg == "" {
		send(bot, msg.Chat.ID, "Usage: /top N (max 100)")
		return
	}

	n, err := strconv.Atoi(arg)
	if err != nil || n <= 0 || n > 100 {
		send(bot, msg.Chat.ID, "N must be between 1 and 100")
		return
	}

	// Здесь будет gRPC вызов

	send(bot, msg.Chat.ID, fmt.Sprintf("Top %d athletes:", n))
}
