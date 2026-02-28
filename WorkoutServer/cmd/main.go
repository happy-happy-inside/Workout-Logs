package main

import (
	"context"
	"log"
	"net"
	hand "workoutserver/internal/server"
	pb "workoutserver/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	logger, _ := zap.NewProduction() // или zap.NewDevelopment() для дев-режима
	defer logger.Sync()

	ctx := context.Background()

	server := hand.NewServer(ctx, logger)
	defer server.Db.Close() //Выполнеться только после завершения

	grpcServer := grpc.NewServer()

	pb.RegisterOrderServiceServer(grpcServer, server)

	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err) // ЖЕСТКО завершает программу без выполнения defer
	}
}
