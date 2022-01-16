package store

import (
	"context"
	"log"

	"github.com/DrGermanius/Shortener/internal/app/models"
	"github.com/DrGermanius/Shortener/internal/store/database"
	"github.com/DrGermanius/Shortener/internal/store/memory"
)

type LinksStorager interface {
	Get(context.Context, string) (string, error)
	GetByUserID(context.Context, string) ([]models.LinkJSON, error)
	Write(context.Context, string, string) (string, error)
	BatchWrite(context.Context, string, []models.BatchOriginal) ([]string, error)
	BatchDelete(ctx context.Context, uid string, links []string) error
	Ping(context.Context) bool
}

func New(connectionString string) (LinksStorager, error) {
	var err error
	var s LinksStorager

	switch connectionString {
	case "":
		s, err = memory.NewLinkMemoryStore()
		if err != nil {
			return nil, err
		}
		log.Println("Service uses inmemory storage")
	default:
		s, err = database.NewDatabaseStore(connectionString)
		if err != nil {
			return nil, err
		}
		log.Println("Service uses database")
	}
	return s, nil
}
