package config

import "strconv"

var c *config

const (
	apiPort = 8080
	host    = "http://localhost"
)

type config struct {
	Port int
	Host string
}

func NewConfig() *config {
	c = new(config)

	c.Port = apiPort
	c.Host = host
	return c
}

func Config() *config {
	return c
}

func (c *config) Full() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}
