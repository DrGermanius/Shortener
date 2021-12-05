package config

import (
	"os"
)

var c *config

const (
	baseURL       = "BASE_URL"
	serverAddress = "SERVER_ADDRESS"
	filePathEnv   = "FILE_STORAGE_PATH"

	defaultFilePath      = "./tmp"
	defaultServerAddress = "localhost:8080"
	defaultBaseURL       = "http://localhost:8080"
)

type config struct {
	BaseURL       string
	ServerAddress string
	FilePath      string
}

func NewConfig() *config {
	c = new(config)

	b, e := os.LookupEnv(serverAddress)
	if !e {
		b = defaultServerAddress
	}
	c.ServerAddress = b

	s, e := os.LookupEnv(baseURL)
	if !e {
		s = defaultBaseURL
	}
	c.BaseURL = s

	p, e := os.LookupEnv(filePathEnv)
	if !e {
		p = defaultFilePath
	}
	c.FilePath = p

	return c
}

func Config() *config {
	return c
}
