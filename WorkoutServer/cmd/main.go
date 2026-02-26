package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 1. Адрес сервера
	serverAddr := "localhost:50051"

	// 2. Устанавливаем соединение с сервером.
	// grpc.WithTransportCredentials(insecure.NewCredentials()) отключает TLS,
	// используйте только для разработки и тестирования.
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer conn.Close() // Важно закрыть соединение, когда оно больше не нужно

	// 3. Создаем клиент для нашего сервиса (генерируется из .proto)
	client := pb.NewGreeterClient(conn)

	// 4. Подготавливаем контекст с таймаутом (рекомендуется для продакшена)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 5. Вызываем удаленный метод
	response, err := client.SayHello(ctx, &pb.HelloRequest{Name: "Мир"})
	if err != nil {
		log.Fatalf("Ошибка при вызове SayHello: %v", err)
	}

	log.Printf("Ответ от сервера: %s", response.GetMessage())
}
