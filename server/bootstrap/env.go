package bootstrap

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Env struct {
	ContextTimeout int
	DBHost         string
	DBPort         string
	DBUser         string
	DBPass         string
	DBName         string
}

func NewEnv() *Env {
	// current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory", err)
	}

	err = godotenv.Load(fmt.Sprintf("%s/.env", dir))
	if err != nil {
		log.Fatalf("Error loading .env file", err)
	}

	return &Env{
		ContextTimeout: strToInt(os.Getenv("CONTEXT_TIMEOUT")),
		DBHost:         os.Getenv("POSTGRES_HOST"),
		DBPort:         os.Getenv("POSTGRES_PORT"),
		DBUser:         os.Getenv("POSTGRES_USER"),
		DBPass:         os.Getenv("POSTGRES_PASSWORD"),
		DBName:         os.Getenv("POSTGRES_DB"),
	}
}

func strToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Error converting string to int")
	}
	return i
}
