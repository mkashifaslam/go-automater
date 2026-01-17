package main

import (
	"log"
	"os"
)

// Env represents the application environment
type Env string

// Possible environments
const (
	Dev  Env = "development"
	Prod Env = "production"
)

func GetEnv() (Env, string) {
	var env Env
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("PORT not set, defaulting to 3000")
		port = "3000"
	}

	env = Env(os.Getenv("ENV"))
	if env == "" {
		log.Println("ENV not set, defaulting to production")
		env = Prod
	}

	log.Printf("Starting server in %s mode on port %s\n", env, port)
	return env, port
}
