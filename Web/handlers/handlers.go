package handlers

import (
	cl "Web/client"
	pb "Web/proto"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MyClient struct {
	Cl *cl.Client
}

func (c *MyClient) Top(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	type topRequest struct {
		Upr   string `json:"upr"`
		Count int64  `json:"count"`
	}

	var req topRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	top, err := c.Cl.TopUsers(ctx, &pb.Uprajnenie{
		Upr:   req.Upr,
		Count: req.Count,
	})
	if err != nil {
		log.Error().Err(err).Msg("gRPC server error")
		http.Error(w, "server ne otvechaet", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(top)
}
func (c *MyClient) AddRes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	type Res struct {
		User string         `json:"user"`
		Upr  []*pb.Podhpowt `json:"upr"`
	}
	var res Res
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		http.Error(w, "sosi", 400)
		return
	}
	addres := pb.AddResRequest{
		User:           res.User,
		SportsExercise: res.Upr,
	}
	response, err := c.Cl.AddRes(ctx, &addres)
	if err != nil {
		log.Error().Err(err).Msg("gRPC server error")
		http.Error(w, "server ne otvechaet", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
func (c *MyClient) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	type Res struct {
		User  string    `json:"user"`
		Upr   []string  `json:"upr"`
		Nach  time.Time `json:"nach"`
		Konec time.Time `json:"konec"`
	}
	var res Res
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		http.Error(w, "sosi", 400)
	}
	nach := timestamppb.New(res.Nach)
	konec := timestamppb.New(res.Konec)
	response, err := c.Cl.GetRes(ctx, &pb.GetResRequest{User: res.User, Upr: res.Upr, Nachalo: nach, Konec: konec})
	if err != nil {
		http.Error(w, "sosi v servere", 400)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
