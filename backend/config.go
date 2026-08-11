package main

import (
	"fmt"
	"os"
)

type Config struct {
	Port string
}

func loadConfig() (Config, error) {
	port := os.Getenv("DJ_BACKEND_PORT")
	if port == "" {
		return Config{}, fmt.Errorf("DJ_BACKEND_PORT env variable is not provided")
	}

	return Config{
		Port: os.Getenv("DJ_BACKEND_PORT"),
	}, nil
}
