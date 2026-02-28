package handlers

import (
	client "bot/client/serverclient"
	"bot/client/aiclient"
	action "bot/internal/botaction"
	"bot/proto"
	pb "bot/proto"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func HandleStart(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := `	 Workout logs 
Это сервис, который позволяет отслеживать тренировочные результаты.
Советую для начала ознакомиться с функционалом для этого введи /help.
Продолжая работу с этим сервсиом ты соглашаешься с тем что все данные вбитые тобой сюда, становиться достоянем общественности`

	if err := action.Send(bot, msg.Chat.ID, text); err != nil {
		log.Print(err)
	}

	if err := action.SendSticker(bot, msg.Chat.ID, "CAACAgIAAxkBAAFDY0xpoeoNsz4tie_zeC7YaZPf88vXmQACbxYAAqqn6UtKk_rNGpoaojoE"); err != nil {
		log.Print(err)
	}
}

func HandleHelp(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := `<b>Список команд:</b>

<b>Приветственное сообщение:</b>
<code>/start</code>

<b>Вывести этот список:</b>
<code>/help</code>

<b>Обрати внимание:</b>
Все что в квадратных <code>[]</code> — опционально.
Все параметры вводятся через пробел.
Если в названии есть пробелы — заменяй их нижними подчеркиваниями.
Пример:
<code>/add присед_со_штангой</code>
━━━━━━━━━━━━━━━━━━━

<b>Добавить новый результат:</b>
<code>/add упражнение повторения подходы вес [дата YYYY-MM-DD] [теги...]</code>
Если не указывать дату — автоматически подставится сегодняшняя (YYYY-MM-DD).
Теги можно писать любые.
Можно указать несколько тегов через пробел до запятой.
━━━━━━━━━━━━━━━━━━━

<b>Управление и поиск результатов:</b>

<b>Поиск:</b>
<code>/get search [upr:упражнение] [data:YYYY-MM-DD] [teg:тег] [reps:повторения]</code>
Выведутся все результаты, совпадающие с указанными параметрами.
Пример:
<code>/get search teg:День_ног</code>

<b>Удаление:</b>
<code>/get del упражнение дата:YYYY-MM-DD</code>

<b>Записи за промежуток:</b>
<code>/get period [дата_от] [дата_до] [упражнение]</code>
Если даты не указаны — от первой записи до текущей даты.
━━━━━━━━━━━━━━━━━━━

<b>Общая статистика:</b>
<code>/stat period дата_от дата_до [упражнение]</code>
Если даты не указаны — от первой записи до текущей даты.

<b>AI-анализ:</b>
<code>/stat AI</code>
Отправляет все записи в LLM для фидбека и рекомендаций.
━━━━━━━━━━━━━━━━━━━

<b>Топ пользователей:</b>

<b>Лучшие:</b>
<code>/top best упражнение n</code>
n — количество людей в топе (1–100).

<b>Мое место:</b>
<code>/top me упражнение</code>`

	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, text)
	msgConfig.ParseMode = "HTML"

	if _, err := bot.Send(msgConfig); err != nil {
		log.Print(err)
	}
}

func HandleAdd(grpcClient *client.Client, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	args := msg.CommandArguments()
	if args == "" {
		log.Printf("HandleAdd: empty args from user %s", msg.From.UserName)
		if err := action.Send(bot, msg.Chat.ID, "Надо так: /add упражнение повторения подходы вес [дата YYYY-MM-DD], ..."); err != nil {
			log.Print(err)
		}
		return
	}

	entries := strings.Split(args, ",")
	var sports []*pb.Podhpowt

	for _, entry := range entries {
		fields := strings.Fields(strings.TrimSpace(entry))

		if len(fields) != 4 && len(fields) != 5 {
			log.Printf("HandleAdd: invalid args count from user %s: %v", msg.From.UserName, fields)
			if err := action.Send(bot, msg.Chat.ID, "Надо так: /add упражнение повторения подходы вес [дата YYYY-MM-DD], ..."); err != nil {
				log.Print(err)
			}
			return
		}

		reps, err1 := strconv.ParseInt(fields[1], 10, 64)
		sets, err2 := strconv.ParseInt(fields[2], 10, 64)
		weight, err3 := strconv.ParseFloat(fields[3], 64)

		if err1 != nil || err2 != nil || err3 != nil {
			log.Printf("HandleAdd: parse error from user %s: %v %v %v", msg.From.UserName, err1, err2, err3)
			if err := action.Send(bot, msg.Chat.ID, "Повторения, подходы и вес должны быть числами"); err != nil {
				log.Print(err)
			}
			return
		}

		var date time.Time
		var err4 error

		if len(fields) == 5 {
			date, err4 = time.Parse("2006-01-02", fields[4])
			if err4 != nil {
				log.Printf("HandleAdd: date parse error from user %s: %v", msg.From.UserName, err4)
				if err := action.Send(bot, msg.Chat.ID, "Напиши дату в формате YYYY-MM-DD"); err != nil {
					log.Print(err)
				}
				return
			}
		} else {
			now := time.Now()
			date = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
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
		User:  msg.From.UserName,
		ToAdd: sports,
	}

	resp, err := grpcClient.AddRes(ctx, req)
	if err != nil {
		log.Printf("HandleAdd: grpc error for user %s: %v", msg.From.UserName, err)
		if err := action.Send(bot, msg.Chat.ID, "Результат не был добавлен. Ошибка:Сервер не отвечает, но мы скоро все починим!"); err != nil {
			log.Print(err)
		}
		return
	}

	if err := action.Send(bot, msg.Chat.ID, resp.Otv); err != nil {
		log.Print(err)
	}
}

func HandleGet(grpcClient *client.Client, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	args := strings.Fields(msg.CommandArguments())

	if len(args) != 1 && len(args) != 3 {
		log.Printf("HandleGet: invalid args from user %s: %v", msg.From.UserName, args)
		if err := action.Send(bot, msg.Chat.ID, "Надо так: /get упражнение [начало YYYY-MM-DD] [конец YYYY-MM-DD]"); err != nil {
			log.Print(err)
		}
		return
	}

	exercise := args[0]

	req := &pb.GetResRequest{
		User: msg.From.UserName,
		Upr:  []string{exercise},
	}

	if len(args) == 3 {
		start, err := time.Parse("2006-01-02", args[1])
		if err != nil {
			log.Printf("HandleGet: start date parse error: %v", err)
			if err := action.Send(bot, msg.Chat.ID, "Начальная дата в формате YYYY-MM-DD"); err != nil {
				log.Print(err)
			}
			return
		}

		end, err := time.Parse("2006-01-02", args[2])
		if err != nil {
			log.Printf("HandleGet: end date parse error: %v", err)
			if err := action.Send(bot, msg.Chat.ID, "Конечная дата в формате YYYY-MM-DD"); err != nil {
				log.Print(err)
			}
			return
		}

		if end.Before(start) {
			log.Printf("HandleGet: end before start from user %s", msg.From.UserName)
			if err := action.Send(bot, msg.Chat.ID, "Конечная дата не может быть раньше начальной"); err != nil {
				log.Print(err)
			}
			return
		}

		req.Nachalo = timestamppb.New(start)
		req.Konec = timestamppb.New(end)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := grpcClient.GetRes(ctx, req)
	if err != nil {
		log.Printf("HandleGet: grpc error for user %s: %v", msg.From.UserName, err)
		if err := action.Send(bot, msg.Chat.ID, " Ошибка:Сервер не отвечает, но мы скоро все починим!"); err != nil {
			log.Print(err)
		}
		return
	}

	if len(resp.Results) == 0 {
		if err := action.Send(bot, msg.Chat.ID, "No results found"); err != nil {
			log.Print(err)
		}
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

	if err := action.Send(bot, msg.Chat.ID, builder.String()); err != nil {
		log.Print(err)
	}
}

func HandleTop(grpcClient *client.Client, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	args := strings.Fields(msg.CommandArguments())
	if len(args) != 2 {
		log.Printf("HandleTop: invalid args from user %s: %v", msg.From.UserName, args)
		if err := action.Send(bot, msg.Chat.ID, "Надо так: /top упражение N (max 100)"); err != nil {
			log.Print(err)
		}
		return
	}

	exercise := args[0]

	n, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil || n <= 0 || n > 100 {
		log.Printf("HandleTop: invalid N from user %s: %v", msg.From.UserName, args[1])
		if err := action.Send(bot, msg.Chat.ID, "N должен быть от 1 до 100"); err != nil {
			log.Print(err)
		}
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
		log.Printf("HandleTop: grpc error for user %s: %v", msg.From.UserName, err)
		if err := action.Send(bot, msg.Chat.ID, " Ошибка:Сервер не отвечает, но мы скоро все починим!"); err != nil {
			log.Print(err)
		}
		return
	}

	if len(resp.Top) == 0 {
		if err := action.Send(bot, msg.Chat.ID, "No data"); err != nil {
			log.Print(err)
		}
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

	if err := action.Send(bot, msg.Chat.ID, builder.String()); err != nil {
		log.Print(err)
	}
}

func HandleStat(grpcClient *client.Client, bot *tgbotapi.BotAPI, msg *tgbotapi.Message, AIclient *aiclient.Client) {
	
	req := &proto.StatRequest {
		Usr: msg.From.UserName,
	}
	

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := grpcClient.Stat(ctx ,req)
	if err != nil {
		if err := action.Send(bot, msg.Chat.ID ,"Ошибка:Сервер не отвечает, но мы скоро все починим!"); err != nil {
			log.Print(err)
		}
		log.Print(err)
	}

	action.Send(aiclient.Get())
}
