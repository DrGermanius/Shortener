package config

import (
	"log"
	"os"
	"strconv"
)

var c *config

const (
	baseURL       = "BASE_URL"
	serverAddress = "SERVER_ADDRESS"

	apiPort = "36595"
	host    = "http://localhost"
)

type config struct {
	Port int
	Host string
}

func NewConfig() *config {
	c = new(config)

	b, e := os.LookupEnv(baseURL)
	if !e {
		b = host
	}
	c.Host = b

	s, e := os.LookupEnv(serverAddress)
	if !e {
		s = apiPort
	}
	port, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalln(err)
	}

	c.Port = port
	return c
}

func Config() *config {
	return c
}

func (c *config) Full() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}
