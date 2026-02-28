package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	pb "aistat/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderServiceServer struct {
	pb.UnimplementedOrderServiceServer
	deepseekAPIKey string
}

func NewOrderServiceServer(apiKey string) *OrderServiceServer {
	return &OrderServiceServer{
		deepseekAPIKey: apiKey,
	}
}

type deepseekRequest struct {
	Model    string            `json:"model"`
	Messages []deepseekMessage `json:"messages"`
}

type deepseekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepseekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (s *OrderServiceServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {

	if req.User == "" {
		return nil, status.Error(codes.InvalidArgument, "user is required")
	}

	var statsText string
	for _, stat := range req.Stat {
		date := ""
		if stat.Date != nil {
			date = stat.Date.AsTime().Format(time.RFC3339)
		}

		statsText += fmt.Sprintf(
			"Упражнение: %s\nВес: %.2f\nПодходы: %d\nПовторения: %d\nДата: %s\n\n",
			stat.Upr,
			stat.Ves,
			stat.Podh,
			stat.Powt,
			date,
		)
	}

	prompt := fmt.Sprintf(
		`Ты выступаешь как профессиональный силовой тренер с опытом программирования тренировок (гипертрофия, сила, рекомпозиция, периодизация).

Я передаю тебе журнал тренировок. 
Проанализируй тренировочную статистику пользователя %s.
Твоя задача — провести глубокий анализ и дать структурированный, конкретный и прикладной фитбек.

Учитывай:
- Объём (подходы на мышечную группу в неделю)
- Интенсивность (рабочий %% от 1ПМ если можно оценить)
- Прогрессию нагрузок
- Баланс тяни/жми, верх/низ, квадрицепс/задняя цепь
- Частоту тренировок
- Повторения в резерве (если можно предположить)
- Возможные признаки перетренированности или недогруза
- Адекватность диапазона повторений под цель
- Логичность структуры микроцикла

Вот журнал тренировок:
%s

Формат ответа:

1. Краткая общая оценка программы (3–6 предложений)
2. Анализ объёма по мышечным группам (таблично или списком)
3. Анализ интенсивности
4. Что сделано хорошо
5. Главные ошибки или слабые места
6. Конкретные рекомендации по корректировке:
   - что убрать
   - что добавить
   - где увеличить вес
   - где уменьшить объём
   - как изменить диапазоны повторений
7. Предложи пример корректировки на следующую неделю
8. Если не хватает данных — задай уточняющие вопросы

Будь конкретным. Не давай общих советов вроде "следи за техникой".
Все рекомендации должны быть обоснованы анализом данных.
`,
		req.User,
		statsText,
	)

	dsReq := deepseekRequest{
		Model: "deepseek-chat",
		Messages: []deepseekMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	body, err := json.Marshal(dsReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.deepseek.com/v1/chat/completions",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.deepseekAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, status.Errorf(codes.Internal, "deepseek error: %s", string(raw))
	}

	var dsResp deepseekResponse
	if err := json.NewDecoder(resp.Body).Decode(&dsResp); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(dsResp.Choices) == 0 {
		return nil, status.Error(codes.Internal, "empty response from deepseek")
	}

	return &pb.GetResponse{
		User:     req.User,
		Response: dsResp.Choices[0].Message.Content,
	}, nil
}
