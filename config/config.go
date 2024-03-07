package config

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	SNOW_USER     string
	SNOW_PASS     string
	SNOW_INSTANCE string
)

func init() {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	// Initialize environment variables
	SNOW_USER = os.Getenv("SNOW_USER")
	SNOW_PASS = os.Getenv("SNOW_PASS")
	SNOW_INSTANCE = os.Getenv("SNOW_INSTANCE")
}
