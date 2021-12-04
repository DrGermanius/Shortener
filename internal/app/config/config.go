package config

import (
	"os"
)

var c *config

const (
	baseURL       = "BASE_URL"
	serverAddress = "SERVER_ADDRESS"

	defaultServerAddress = "localhost:8080"
	defaultBaseUrl       = "http://localhost:8080/"
)

type config struct {
	BaseUrl       string
	ServerAddress string
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
		s = defaultBaseUrl
	}

	c.BaseUrl = s
	return c
}

func Config() *config {
	return c
}
