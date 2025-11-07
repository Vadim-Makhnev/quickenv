package main

import (
	"log"

	"github.com/Vadim-Makhnev/quickenv"
)

func main() {
	// Load .env file with debug output
	count, err := quickenv.Load(&quickenv.LoadOptions{
		Pathname:  ".env.example", // Name of the env file to load
		Debug:     true,           // Print loaded/skipped lines
		MaxLevels: 3,              // Search up to 3 parent directories
	})
	if err != nil {
		log.Fatal("Failed to load .env:", err)
	}

	// Safely get values with defaults
	configPath := quickenv.GetEnv("CONFIG_PATH", "config/local.yaml")
	dbPort := quickenv.GetEnv("DB_PORT", "8000")

	log.Printf("Loaded %d vars\n", count)
	log.Println("CONFIG_PATH =", configPath)
	log.Println("DB_PORT     =", dbPort)
}
