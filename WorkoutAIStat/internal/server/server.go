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
)

type OrderServiceServer struct {
	deepseekAPIKey string
}

func NewOrderServiceServer(apiKey string) *OrderServiceServer {
	return &OrderServiceServer{
		deepseekAPIKey: apiKey,
	}
}

func (s *OrderServiceServer) Handle(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {

	if req.User == "" {
		return nil, fmt.Errorf("user is required")
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

	prompt := fmt.Sprintf(`Ты выступаешь как профессиональный силовой тренер с опытом программирования тренировок (гипертрофия, сила, рекомпозиция, периодизация).\
	Я передаю тебе журнал тренировок. Проанализируй тренировочную статистику пользователя %s.
	Твоя задача — провести глубокий анализ и дать структурированный, конкретный и прикладной фитбек.
	Учитывай: 
		- Объём (подходы на мышечную группу в неделю) 
		- Интенсивность (рабочий %% от 1ПМ если можно оценить) 
		- Прогрессию нагрузок 
		- Баланс тяни/жми, верх/низ, квадрицепс/задняя цепь 
		- Частоту тренировок - Повторения в резерве (если можно предположить) 
		- Возможные признаки перетренированности или недогруза - Адекватность диапазона повторений под цель 
		- Логичность структуры микроцикла 
	Вот журнал тренировок: 
		%s 
	Формат ответа:
	Отвечай только текстом, не от чего, просто рицензия. Не используй табличики и прочее, только текст, по следующемо шаблону:
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
	Не давай общих советов вроде "следи за техникой". Все рекомендации должны быть обоснованы анализом данных и написанны на русском.`,
		req.User,
		statsText,
	)

	dsReq := map[string]interface{}{
		"model": "meta-llama/llama-3-8b-instruct",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	body, err := json.Marshal(dsReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://openrouter.ai/api/v1/chat/completions",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.deepseekAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openrouter error: %s", string(raw))
	}

	var dsResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&dsResp); err != nil {
		return nil, err
	}

	if len(dsResp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from openrouter")
	}

	return &pb.GetResponse{
		Reqid:    req.Reqid,
		Response: dsResp.Choices[0].Message.Content,
	}, nil
}
