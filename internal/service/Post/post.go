package Post

import (
	"database/sql"
	"fmt"
	"log"
	"main/internal/storage/db"
	"main/pkg/PostAPIService"
	"os"
)

type Server struct {
	PostAPIService.UnimplementedPostAPIServer
	conn  *sql.DB
	query *db.Queries
}

func getQuery() (*db.Queries, *sql.DB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, nil, err
	}
	q := db.New(conn)
	return q, conn, nil
}

func NewServer() *Server {
	q, conn, err := getQuery()
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	return &Server{
		conn:  conn,
		query: q,
	}
}
