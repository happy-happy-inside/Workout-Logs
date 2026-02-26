package structures

import (
	pb "ai-saas/proto"
	"context"
	"fmt"

	db "ai-saas/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Server struct {
	must   pb.UnimplementedOrderServiceServer
	db     pgxpool.Pool
	logger zap.Logger
}

type GetResModel struct {
	ves  float64
	podh int
	powt int
}
type top struct {
	user string
	ves  float64
}

func NewServer(logger zap.Logger, ctx context.Context) *Server {
	return &Server{
		must:   pb.UnimplementedOrderServiceServer{},
		db:     *db.NewPool(ctx),
		logger: logger,
	}
}
func (s *Server) AddRes(ctx context.Context, req *pb.AddResRequest) (*pb.AddResResponse, error) {
	batch := pgx.Batch{}
	for _, v := range req.SportsExercise {
		batch.Queue(`INSERT INTO KACH (USER,UPR,VES,PODH,POWT,DATE) VALUES($1,$2,$3,$4,$5,$6)`, req.User, v.Upr, v.Ves, v.Podh, v.Powt, v.Date)
	}
	br := s.db.SendBatch(ctx, &batch)
	defer br.Close()

	// Обязательно нужно вычитать все результаты!
	for i := 0; i < len(req.SportsExercise); i++ {
		_, err := br.Exec()
		if err != nil {
			s.logger.Error("ebat hendler AddRes %v", zap.Error(err))
			return &pb.AddResResponse{Otv: "vse NE zaebis"}, nil
		}
	}

	fmt.Println("AddRes выполнен")
	return &pb.AddResResponse{Otv: "vse zaebis"}, nil
}
func (s *Server) GetRes(ctx context.Context, req *pb.GetResRequest) (*pb.GetResResponse, error) {
	grpcRes := &pb.GetResResponse{}
	for i := range req.Upr {
		res, err := s.db.Query(ctx, `SELECT (PODH,POWT,VES) FROM KACH WHERE DATE>=$1 AND DATE<=$2 AND UPR = '$3' ORDER BY DATE ASC`, req.Nachalo.AsTime(), req.Konec.AsTime(), req.Upr[i])
		if err != nil {
			s.logger.Error("error in GetRes", zap.Error(err))
			return &pb.GetResResponse{}, err
		}
		defer res.Close()
		var models []GetResModel
		j := 0
		max := 0.0
		for res.Next() {
			var model GetResModel
			err := res.Scan(&model.podh, &model.powt, &model.ves)
			if err != nil {
				s.logger.Error("ne mogy rows scan v GetRes", zap.Error(err))
				return &pb.GetResResponse{}, err
			}
			if model.ves > max {
				max = model.ves
			}
			models = append(models, model)
			j++
		}
		raznica := max - models[0].ves
		sr := raznica / float64(len(models))
		grpcRes.Results = append(grpcRes.Results, &pb.Get{Upr: req.Upr[i], Slab: models[0].ves, Siln: max, Sr: sr, Raznica: raznica})
	}
	return grpcRes, nil
}
func (s *Server) TopUsers(ctx context.Context, req *pb.Uprajnenie) (*pb.Top, error) {
	var tops []*pb.Dinah
	res, err := s.db.Query(ctx, `SELECT (USER) FROM KACH WHERE UPR=$1 ORDER BY DESK LIMIT $2`, req.Upr, req.Count)
	if err != nil {
		s.logger.Error("error in TOP query", zap.Error(err))
		return &pb.Top{}, err
	}
	defer res.Close()
	for res.Next() {
		var t pb.Dinah
		err = res.Scan(&t.User, &t.Ves)
		if err != nil {
			s.logger.Error("error in TOP scan", zap.Error(err))
			return &pb.Top{}, err
		}
		tops = append(tops, &t)
	}
	return &pb.Top{Top: tops}, nil
}
