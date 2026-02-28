package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	pb "aistat/proto"
	service "aistat/internal/server"
)

func main() {

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("DEEPSEEK_API_KEY not set")
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	orderService := service.NewOrderServiceServer(apiKey)

	pb.RegisterOrderServiceServer(grpcServer, orderService)

	log.Println("gRPC server started on :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}