package config

import "strconv"

var c *config

const (
	apiPort = 8080
	host    = "http://localhost"
)

type config struct {
	port int
	host string
}

func NewConfig() *config {
	c = new(config)

	c.port = apiPort
	c.host = host
	return c
}

func Config() *config {
	return c
}

func (c *config) Port() int {
	return c.port
}

func (c *config) Host() string {
	return c.host
}

func (c *config) Full() string {
	return c.host + ":" + strconv.Itoa(c.port)
}
