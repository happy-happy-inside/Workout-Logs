package main

import (
	"bot/client"
	pb "bot/proto"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	grpcClient, err := client.NewClient(os.Getenv("HOST"))
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		go handleMessage(grpcClient, bot, update.Message)
	}
}

func handleMessage(grpcClient *client.Client, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	switch msg.Command() {

	case "start":
		handleStart(bot, msg)

	case "help":
		handleHelp(bot, msg)

	case "add":
		handleAdd(grpcClient, bot, msg)

	case "get":
		handleGet(grpcClient, bot, msg)

	case "top":
		handleTop(grpcClient, bot, msg)

	default:
		send(bot, msg.Chat.ID, "Неизвестная команда. Используй /help")
	}
}

func send(bot *tgbotapi.BotAPI, chatID int64, text string) {
	message := tgbotapi.NewMessage(chatID, text)
	bot.Send(message)
}

func sendSticker(bot *tgbotapi.BotAPI, chatID int64, fileID string) error {
	sticker := tgbotapi.NewSticker(chatID, tgbotapi.FileID(fileID))
	_, err := bot.Send(sticker)
	return err
}

func handleStart(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := `<--| Workout logs |-->
Это сервис, который позволяет отслеживать тренировочные результаты.
Чтобы ознакомиться с функционалом — введи /help.`

	send(bot, msg.Chat.ID, text)
}

func handleHelp(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := `Список команд:
Приветсвенное сообщение:
/start

Вывести этот список:
/help

Добавить новый результат:
/add упражнение повторения подходы вес дата,формат:YYYY-MM-DD, ...

Вывести результат:
/get упражнение [начало YYYY-MM-DD] [конец YYYY-MM-DD]

Вывести топ пользователей:
/top упражнение n (n - кол-во людей в топе)`

	send(bot, msg.Chat.ID, text)
}

func handleAdd(grpcClient *client.Client, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	args := msg.CommandArguments()
	if args == "" {
		send(bot, msg.Chat.ID, "Надо так: /add упражнение повторения подходы вес дата,формат:YYYY-MM-DD, ...")
		return
	}

	entries := strings.Split(args, ",")
	var sports []*pb.Podhpowt

	for _, entry := range entries {
		fields := strings.Fields(strings.TrimSpace(entry))
		if len(fields) != 5 {
			send(bot, msg.Chat.ID, "Надо так: /add упражнение повторения подходы вес дата,формат:YYYY-MM-DD, ...")
			return
		}

		reps, err1 := strconv.ParseInt(fields[1], 10, 64)
		sets, err2 := strconv.ParseInt(fields[2], 10, 64)
		weight, err3 := strconv.ParseFloat(fields[3], 64)
		date, err4 := time.Parse("2006-01-02", fields[4])

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			send(bot, msg.Chat.ID, "Напиши дату в формате (YYYY-MM-DD)")
			return
		}

		sports = append(sports, &pb.Podhpowt{
			Upr:  fields[0],
			Ves:  weight,
			Podh: sets,
			Powt: reps,
			Date: timestamppb.New(date),
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.AddResRequest{
		User:           msg.From.UserName,
		SportsExercise: sports,
	}

	resp, err := grpcClient.AddRes(ctx, req)
	if err != nil {
		log.Print(err)
		send(bot, msg.Chat.ID, "Ошибка! Результат не был добавлен,проблема с сервером,  подожди чуток и попробуй еще")
		return
	}

	send(bot, msg.Chat.ID, resp.Otv)
}

func handleGet(grpcClient *client.Client, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	args := strings.Fields(msg.CommandArguments())
	if len(args) == 0 {
		send(bot, msg.Chat.ID, "Надо так: /get упражнение [начало YYYY-MM-DD] [конец YYYY-MM-DD]")
		return
	}

	exercise := args[0]

	req := &pb.GetResRequest{
		User: msg.From.UserName,
		Upr:  []string{exercise},
	}

	if len(args) == 3 {
		start, err1 := time.Parse("2006-01-02", args[1])
		end, err2 := time.Parse("2006-01-02", args[2])
		if err1 != nil || err2 != nil {
			send(bot, msg.Chat.ID, "Напиши дату в формате (YYYY-MM-DD)")
			return
		}

		req.Nachalo = timestamppb.New(start)
		req.Konec = timestamppb.New(end)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := grpcClient.GetRes(ctx, req)
	if err != nil {
		log.Print(err)
		send(bot, msg.Chat.ID, "Ошибка! Результат не удалось найти,проблема с сервером, подожди чуток и попробуй еще")
		return
	}

	if len(resp.Results) == 0 {
		send(bot, msg.Chat.ID, "No results found")
		return
	}

	var builder strings.Builder
	builder.WriteString("Results:\n")

	for _, r := range resp.Results {
		builder.WriteString(fmt.Sprintf(
			"%s | min: %.2f | max: %.2f | avg: %.2f | diff: %.2f\n",
			r.Upr, r.Slab, r.Siln, r.Sr, r.Raznica,
		))
	}

	send(bot, msg.Chat.ID, builder.String())
}

func handleTop(grpcClient *client.Client, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	args := strings.Fields(msg.CommandArguments())
	if len(args) != 2 {
		send(bot, msg.Chat.ID, "Usage: /top упражение N (max 100)")
		return
	}

	exercise := args[0]

	n, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil || n <= 0 || n > 100 {
		send(bot, msg.Chat.ID, "N must be between 1 and 100")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.Uprajnenie{
		Upr:   exercise,
		Count: n,
	}

	resp, err := grpcClient.TopUsers(ctx, req)
	if err != nil {
		log.Print(err)
		send(bot, msg.Chat.ID, "Failed to get top users")
		return
	}

	if len(resp.Top) == 0 {
		send(bot, msg.Chat.ID, "No data")
		return
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Top %d for %s:\n", n, exercise))

	for i, u := range resp.Top {
		builder.WriteString(fmt.Sprintf(
			"%d. %s — %.2f kg\n",
			i+1, u.User, u.Ves,
		))
	}

	send(bot, msg.Chat.ID, builder.String())
}
