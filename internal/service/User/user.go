package User

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"main/pkg/UserAPIService"
	"os"
)

type Server struct {
	UserAPIService.UnimplementedUserAPIServer
	pool *pgxpool.Pool
	//querier *db.Queries
	hmacSecret []byte
}

func getPool() (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	return pgxpool.New(context.Background(), connStr)
}

func NewServer() *Server {
	pool, err := getPool()
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	return &Server{
		pool: pool,
		//querier: db.New(pool),
		hmacSecret: []byte(os.Getenv("hmacSecret")),
	}
}
