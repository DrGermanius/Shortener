package database

import (
	"context"
	"github.com/DrGermanius/Shortener/internal/app/models"

	"github.com/jackc/pgx/v4"
)

type DB struct {
	conn *pgx.Conn
}

func NewDatabaseStorage(connString string) (*DB, error) {
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, err
	}
	return &DB{conn: conn}, nil
}

func (d *DB) Get(string) (string, bool) {
	return "", false
}

func (d *DB) GetByUserID(id string) []models.LinkJSON {
	return []models.LinkJSON{}
}

func (d *DB) Write(uuid, long string) (string, error) {
	return "", nil
}

func (d *DB) Ping() bool {
	return false
}
