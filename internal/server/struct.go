package structures

import (
	pb "ai-saas/proto"
	"context"

	db "ai-saas/internal/storage"

	pgx "github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Server struct {
	must   pb.UnimplementedOrderServiceServer
	db     pgx.Pool
	logger zap.Logger
}

func NewServer(logger zap.Logger, ctx context.Context) *Server {
	return &Server{
		must:   pb.UnimplementedOrderServiceServer{},
		db:     *db.NewPool(ctx),
		logger: logger,
	}
}
func (s *Server) AddRes(ctx context.Context, req *pb.AddResRequest) (*pb.AddResResponse, error) {
	for _, v := range req.SportsExercise {
		_, err := s.db.Exec(ctx, d)
	}
}
