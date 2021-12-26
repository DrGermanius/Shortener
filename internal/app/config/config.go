package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
)

var cfg *config

const (
	baseURL            = "BASE_URL"
	serverAddress      = "SERVER_ADDRESS"
	filePathEnv        = "FILE_STORAGE_PATH"
	dbConnectionString = "DATABASE_DSN"

	defaultFilePath      = "./tmp"
	defaultServerAddress = "localhost:8080"
	defaultBaseURL       = "http://localhost:8080"
)

type config struct {
	BaseURL          string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	ServerAddress    string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	FilePath         string `env:"FILE_STORAGE_PATH" envDefault:"./tmp"`
	ConnectionString string `env:"DATABASE_DSN" envDefault:""`
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "12345"
)

func NewConfig() (*config, error) {
	cfg = &config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	conn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s sslmode=disable",
		host, port, user, password)

	flag.StringVar(&cfg.ServerAddress, "h", cfg.ServerAddress, "host to listen on")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "baseURl for short link")
	flag.StringVar(&cfg.FilePath, "f", cfg.FilePath, "filePath for links")
	flag.StringVar(&cfg.ConnectionString, "d", conn, "postgres connection path")
	flag.Parse()

	return cfg, nil
}

func TestConfig() (*config, error) {
	cfg = &config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func Config() *config {
	return cfg
}
