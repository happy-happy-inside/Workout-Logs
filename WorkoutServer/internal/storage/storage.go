package storage

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	pgx "github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context) *pgx.Pool {

	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("fale connect to DB: %v\n", err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("Net podkluch: %v\n", err)
	}
	fmt.Println("DB work")
	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS KACH(
	ID SERIAL PRIMARY KEY,
	USERNAME TEXT NOT NULL,
	UPR TEXT NOT NULL,
	VES REAL NOT NULL,
	PODH INT NOT NULL,
	POWT INT NOT NULL,
	DATE DATE NOT NULL) `)
	if err != nil {
		log.Fatalf("suka ne mogu sozdat tablicu")
	}
	return pool
}
