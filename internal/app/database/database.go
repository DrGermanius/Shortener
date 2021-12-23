package database

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type DB struct {
	conn *pgx.Conn
}

func CreateConnection(connString string) (*DB, error) {
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, err
	}
	return &DB{conn: conn}, nil
}
