package main

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	ServerName     string `env:"SERVER_NAME"`
	DatabaseName   int    `env:"DATABASE_NAME"`
	SecretKey      string `env:"SECRET_KEY"`
	SignupPassword string `env:"SIGNUP_PASSWORD"`
	DBName         string `env:"DB_NAME"`
	// SignupPasswordRequired bool
	// UsingCustomDBPath bool
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Couldn't load .env", err)
	}

	var config Config
	if err := env.Parse(&config); err != nil {
		log.Fatalf("Couldn't read .env: %v", err)
		return
	}
}
