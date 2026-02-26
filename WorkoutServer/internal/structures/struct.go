package structures

import (
	pb "ai-saas/proto"

	pgx "github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Server struct {
	hm     pb.UnimplementedOrderServiceServer
	db     pgx.Pool
	logger zap.Logger
}

func NewServer(logger zap.Logger, pool pgx.Pool) *Server {

}
