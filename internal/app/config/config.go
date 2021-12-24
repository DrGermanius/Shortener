package config

import (
	"flag"
	"fmt"
	"os"
)

var c *config

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
	BaseURL          string
	ServerAddress    string
	FilePath         string
	ConnectionString string
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "12345"
)

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

	conn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s sslmode=disable",
		host, port, user, password)

	d, e := os.LookupEnv(dbConnectionString)
	if !e {
		d = conn
	}
	flag.StringVar(&d, "d", d, "postgres connection path")
	flag.Parse()

	c.ServerAddress = a
	c.BaseURL = b
	c.FilePath = f
	c.ConnectionString = d

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
	c.ConnectionString = ""

	return c
}

func Config() *config {
	return c
}
