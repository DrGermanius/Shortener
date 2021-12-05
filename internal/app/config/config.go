package config

import (
	"flag"
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

	a, e := os.LookupEnv(serverAddress)
	if !e {
		a = defaultServerAddress
	}
	flag.StringVar(&a, "h", a, "host to listen on")

	b, e := os.LookupEnv(baseURL)
	if !e {
		b = defaultBaseURL
	}
	flag.StringVar(&b, "b", b, "baseURl for short link")

	f, e := os.LookupEnv(filePathEnv)
	if !e {
		f = defaultFilePath
	}
	flag.StringVar(&f, "f", f, "filePath for links")
	flag.Parse()

	c.ServerAddress = a
	c.BaseURL = b
	c.FilePath = f

	return c
}

func Suite() *config { //todo rewrite after lesson about TestSuite
	c = new(config)

	a, e := os.LookupEnv(serverAddress)
	if !e {
		a = defaultServerAddress
	}

	b, e := os.LookupEnv(baseURL)
	if !e {
		b = defaultBaseURL
	}

	f, e := os.LookupEnv(filePathEnv)
	if !e {
		f = defaultFilePath
	}

	c.ServerAddress = a
	c.BaseURL = b
	c.FilePath = f

	return c
}

func Config() *config {
	return c
}
