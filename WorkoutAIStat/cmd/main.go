package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"

	service "aistat/internal/server"
	pb "aistat/proto"

	"google.golang.org/protobuf/proto"
)

func main() {

	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		log.Fatal("AI_API_KEY not set")
	}

	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "kafka:9092"
	}

	svc := service.NewOrderServiceServer(apiKey)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   "ai.requests",
		GroupID: "ai-group",
	})

	writer := &kafka.Writer{
		Addr:  kafka.TCP(broker),
		Topic: "ai.responses",
	}

	log.Println("AI Kafka consumer started")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Println("reader stopped:", err)
			break
		}

		var req pb.GetRequest
		if err := proto.Unmarshal(msg.Value, &req); err != nil {
			log.Println("unmarshal error:", err)
			continue
		}

		resp, err := svc.Handle(ctx, &req)

		if err != nil {
			resp = &pb.GetResponse{
				Reqid: req.Reqid,
				Error: err.Error(),
			}
		}

		data, _ := proto.Marshal(resp)

		err = writer.WriteMessages(ctx, kafka.Message{
			Key:   msg.Key,
			Value: data,
		})
		if err != nil {
			log.Println("write response error:", err)
		}
	}
}
