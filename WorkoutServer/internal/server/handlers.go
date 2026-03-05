package structures

import (
	"context"
	"fmt"
	"strconv"
	"time"
	pb "workoutserver/proto"

	db "workoutserver/internal/storage"

	redis "workoutserver/internal/redis"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedOrderServiceServer
	Db     pgxpool.Pool
	Logger *zap.Logger
	Redis  *redis.RedisClient
}

type GetResModel struct {
	ves  float64
	podh int
	powt int
}

func NewServer(ctx context.Context, logger *zap.Logger, redis *redis.RedisClient) *Server {
	return &Server{
		Db:     *db.NewPool(ctx),
		Logger: logger,
		Redis:  redis,
	}
}
func (s *Server) AddRes(ctx context.Context, req *pb.AddResRequest) (*pb.AddResResponse, error) {
	batch := pgx.Batch{}
	for _, v := range req.ToAdd {
		date := v.Date.AsTime()
		batch.Queue(`INSERT INTO KACH (USERNAME,UPR,VES,PODH,POWT,DATE) VALUES($1,$2,$3,$4,$5,$6)`, req.User, v.Upr, v.Ves, v.Podh, v.Powt, date)
	}
	br := s.Db.SendBatch(ctx, &batch)
	defer br.Close()

	// Обязательно нужно вычитать все результаты!
	for i := 0; i < len(req.ToAdd); i++ {
		_, err := br.Exec()
		if err != nil {
			s.Logger.Error("ebat hendler AddRes %v", zap.Error(err))
			return &pb.AddResResponse{Otv: "vse NE zaebis"}, nil
		}
	}

	fmt.Println("AddRes выполнен")
	return &pb.AddResResponse{Otv: "vse zaebis"}, nil
}
func (s *Server) GetRes(ctx context.Context, req *pb.GetResRequest) (*pb.GetResResponse, error) {
	grpcRes := &pb.GetResResponse{}
	for i := range req.Upr {
		res, err := s.Db.Query(ctx, `SELECT PODH,POWT,VES FROM KACH WHERE DATE>=$1 AND DATE<=$2 AND UPR = $3 ORDER BY DATE ASC`, req.Nachalo.AsTime(), req.Konec.AsTime(), req.Upr[i])
		if err != nil {
			s.Logger.Error("error in GetRes", zap.Error(err))
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
				s.Logger.Error("ne mogy rows scan v GetRes", zap.Error(err))
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

	key := req.Upr + strconv.Itoa(int(req.Count))
	data, err := s.Redis.Get(ctx, key)
	if nil == err {
		var top pb.Top

		if err := proto.Unmarshal(data, &top); nil == err {
			return &top, nil
		}
	}

	var tops []*pb.Dinah
	res, err := s.Db.Query(ctx, `SELECT USERNAME,VES FROM KACH WHERE UPR=$1 ORDER BY DATE DESC LIMIT $2`, req.Upr, req.Count)
	if err != nil {
		s.Logger.Error("error in TOP query", zap.Error(err))
		return &pb.Top{}, err
	}
	defer res.Close()
	for res.Next() {
		var t pb.Dinah
		err = res.Scan(&t.User, &t.Ves)
		if err != nil {
			s.Logger.Error("error in TOP scan", zap.Error(err))
			return &pb.Top{}, err
		}
		tops = append(tops, &t)
	}

	Top := &pb.Top{Top: tops}

	data, protoErr := proto.Marshal(Top)
	if nil == protoErr {
		s.Redis.Set(ctx, key, data, 1*time.Hour)
	} else {
		s.Logger.Error("can`t marshal cahse:", zap.Error(protoErr))
	}

	return Top, nil
}

func (s *Server) Stat(ctx context.Context, req *pb.StatRequest) (*pb.StatResponse, error) {
	if req.Usr == "" {
		return &pb.StatResponse{}, fmt.Errorf("empty user")
	}

	rows, err := s.Db.Query(ctx,
		`SELECT UPR, VES, PODH, POWT, DATE 
		 FROM KACH 
		 WHERE USERNAME=$1 
		 ORDER BY DATE ASC`,
		req.Usr,
	)
	if err != nil {
		s.Logger.Error("error in Stat query", zap.Error(err))
		return &pb.StatResponse{}, err
	}
	defer rows.Close()

	resp := &pb.StatResponse{
		User: req.Usr,
	}

	for rows.Next() {
		var (
			upr  string
			ves  float32
			podh int32
			powt int32
			date time.Time
		)

		err := rows.Scan(&upr, &ves, &podh, &powt, &date)
		if err != nil {
			s.Logger.Error("error in Stat scan", zap.Error(err))
			return &pb.StatResponse{}, err
		}

		resp.Stat = append(resp.Stat, &pb.Podhpowt{
			Upr:  upr,
			Ves:  float64(ves),
			Podh: int64(podh),
			Powt: int64(powt),
			Date: timestamppb.New(date),
		})
	}

	if rows.Err() != nil {
		s.Logger.Error("rows error in Stat", zap.Error(rows.Err()))
		return &pb.StatResponse{}, rows.Err()
	}

	return resp, nil
}
