package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL       string
	GoogleClientID    string
	GoogleClientSecret string
	CallbackURL       string
	SessionSecret     string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// Continue if .env doesn't exist
	}

	// Get the host from environment or default to localhost:3000
	host := os.Getenv("HOST")
	if host == "" {
		host = "http://localhost:3000"
	}

	// Get session secret or generate a default one
	sessionSecret := os.Getenv("SESSION_SECRET")
	// if sessionSecret == "" {
	// 	sessionSecret = "your-secret-key-minimum-32-chars-long"
	// }

	// Construct callback URL to match Google OAuth setting
	callbackURL := fmt.Sprintf("%s/callback", host)

	config := &Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		CallbackURL:        callbackURL,
		SessionSecret:      sessionSecret,
	}

	// Log the callback URL for debugging
	fmt.Printf("Callback URL configured as: %s\n", callbackURL)

	return config, nil
}
