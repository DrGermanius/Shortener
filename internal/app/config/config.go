package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

var c *config

const (
	baseURL            = "BASE_URL"
	serverAddress      = "SERVER_ADDRESS"
	filePathEnv        = "FILE_STORAGE_PATH"
	dbConnectionString = "DATABASE_DSN"
	authKey            = "AUTH_KEY"
	workersCount       = "WORKERS_COUNT"
	jsonConfig         = "CONFIG"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "12345"
)

var (
	defaultFilePath      = "./tmp"
	defaultServerAddress = "localhost:8080"
	defaultBaseURL       = "http://localhost:8080"
	defaultAuthKey       = "secret"
	defaultWorkersCount  = "10"
)

type config struct {
	BaseURL          string `json:"base_url"`
	ServerAddress    string `json:"server_address"`
	FilePath         string `json:"file_storage_path"`
	ConnectionString string `json:"database_dsn"`
	AuthKey          string `json:"auth_key"`
	WorkersCount     string `json:"workers_count"`
	IsHTTPS          bool   `json:"enable_https"`
}

func NewConfig() (*config, error) {
	c = new(config)

	defaultConn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s sslmode=disable",
		host, port, user, password)

	var jc string
	flag.StringVar(&jc, "c", setEnvOrDefault(jsonConfig, ""), "json config")
	if jc != "" {
		jsConf, err := readConfigFromJSON(jc)
		if err != nil {
			return nil, err
		}
		if jsConf.ServerAddress != "" {
			defaultServerAddress = jsConf.ServerAddress
		}
		if jsConf.AuthKey != "" {
			defaultAuthKey = jsConf.AuthKey
		}
		if jsConf.BaseURL != "" {
			defaultBaseURL = jsConf.BaseURL
		}
		if jsConf.FilePath != "" {
			defaultFilePath = jsConf.FilePath

		}
		if jsConf.ConnectionString != "" {
			defaultConn = jsConf.ConnectionString
		}
	}

	c.AuthKey = setEnvOrDefault(authKey, defaultAuthKey)
	c.WorkersCount = setEnvOrDefault(workersCount, defaultWorkersCount)
	flag.StringVar(&c.ServerAddress, "h", setEnvOrDefault(serverAddress, defaultServerAddress), "host to listen on")
	flag.StringVar(&c.BaseURL, "b", setEnvOrDefault(baseURL, defaultBaseURL), "baseURl for short link")
	flag.StringVar(&c.FilePath, "f", setEnvOrDefault(filePathEnv, defaultFilePath), "filePath for links")
	flag.StringVar(&c.ConnectionString, "d", setEnvOrDefault(dbConnectionString, defaultConn), "postgres connection path")
	flag.Parse()
	c.IsHTTPS = isFlagPassed("s")
	return c, nil
}

func SetTestConfig() *config {
	c = new(config)
	c.AuthKey = setEnvOrDefault(authKey, defaultAuthKey)
	c.ServerAddress = setEnvOrDefault(serverAddress, defaultServerAddress)
	c.BaseURL = setEnvOrDefault(baseURL, defaultBaseURL)
	c.FilePath = setEnvOrDefault(filePathEnv, defaultFilePath)
	c.ConnectionString = setEnvOrDefault(dbConnectionString, "")
	return c
}

func Config() *config {
	return c
}

func readConfigFromJSON(path string) (*config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	conf := new(config)
	err = json.Unmarshal(b, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func setEnvOrDefault(env, def string) string {
	res, e := os.LookupEnv(env)
	if !e {
		res = def
	}
	return res
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
