package bootstrap

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Env struct {
	ServerAddress  string
	Port           string
	ContextTimeout int
	DBHost         string
	DBPort         string
	DBUser         string
	DBPass         string
	DBName         string
}

func NewEnv() (*Env, error) {
	// current working directory
	dir, err := os.Getwd()
	if err != nil {
		// logger.Errorf(nil, "Error getting current working directory: %v", err)
		return nil, err
	}

	// logger.Info(nil, "loading .env")
	err = godotenv.Load(fmt.Sprintf("%s/.env", dir))
	if err != nil {
		// logger.Errorf(nil, "Error loading .env file: %v", err)
		return nil, err
	}

	timeout, err := strToInt(os.Getenv("CONTEXT_TIMEOUT"))
	if err != nil {
		// logger.Errorf(nil, "Error converting string to int: %v", err)
		return nil, err
	}

	return &Env{
		ServerAddress:  os.Getenv("SERVER_ADDRESS"),
		Port:           os.Getenv("PORT"),
		ContextTimeout: timeout,
		DBHost:         os.Getenv("POSTGRES_HOST"),
		DBPort:         os.Getenv("POSTGRES_PORT"),
		DBUser:         os.Getenv("POSTGRES_USER"),
		DBPass:         os.Getenv("POSTGRES_PASSWORD"),
		DBName:         os.Getenv("POSTGRES_DB"),
	}, nil
}

func strToInt(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		// logger.Errorf(nil, fmt.Sprintf("Error converting string to int: %v"), err)
		return 0, err
	}
	return i, nil
}
