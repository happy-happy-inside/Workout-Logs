package main

import (
	cl "Web/client"
	hand "Web/handlers"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	c, err := cl.NewClient(os.Getenv("CLIENT_ADDR"))
	if err != nil {
		log.Error().Err(err)
	}
	defer c.Close()
	h := hand.MyClient{Cl: c}

	r.Post("/top", h.Top)
	r.Post("/add", h.AddRes)
	r.Post("/getres", h.Get)
	http.ListenAndServe(":8080", r)
}
