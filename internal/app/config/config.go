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
	authKey            = "AUTH_KEY"

	defaultFilePath      = "./tmp"
	defaultServerAddress = "localhost:8080"
	defaultBaseURL       = "http://localhost:8080"
	defaultAuthKey       = "secret"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "12345"
)

type config struct {
	BaseURL          string
	ServerAddress    string
	FilePath         string
	ConnectionString string
	AuthKey          string
}

func NewConfig() *config {
	c = new(config)

	defaultConn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s sslmode=disable",
		host, port, user, password)

	c.AuthKey = setEnvOrDefault(authKey, defaultAuthKey)
	flag.StringVar(&c.ServerAddress, "h", setEnvOrDefault(serverAddress, defaultServerAddress), "host to listen on")
	flag.StringVar(&c.BaseURL, "b", setEnvOrDefault(baseURL, defaultBaseURL), "baseURl for short link")
	flag.StringVar(&c.FilePath, "f", setEnvOrDefault(filePathEnv, defaultFilePath), "filePath for links")
	flag.StringVar(&c.ConnectionString, "d", setEnvOrDefault(dbConnectionString, defaultConn), "postgres connection path")
	flag.Parse()

	return c
}

func SetTestConfig() *config {
	c = new(config)
	c.ServerAddress = setEnvOrDefault(serverAddress, defaultServerAddress)
	c.BaseURL = setEnvOrDefault(baseURL, defaultBaseURL)
	c.FilePath = setEnvOrDefault(filePathEnv, defaultFilePath)
	c.ConnectionString = setEnvOrDefault(dbConnectionString, "")
	return c
}

func Config() *config {
	return c
}

func setEnvOrDefault(env, def string) string {
	res, e := os.LookupEnv(env)
	if !e {
		res = def
	}
	return res
}
