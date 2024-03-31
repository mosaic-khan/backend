package User

import (
	"database/sql"
	"fmt"
	"log"
	"main/internal/storage/db"
	"main/pkg/UserAPIService"
	"os"

	_ "github.com/lib/pq"
)

type Server struct {
	UserAPIService.UnimplementedUserAPIServer
	query      *db.Queries
	hmacSecret []byte
}

func getQuery() (*db.Queries, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	conn, err := sql.Open("postges", connStr)
	if err != nil {
		return nil, err
	}
	q := db.New(conn)
	return q, nil
}

func NewServer() *Server {
	q, err := getQuery()
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	return &Server{
		query:      q,
		hmacSecret: []byte(os.Getenv("hmacSecret")),
	}
}
