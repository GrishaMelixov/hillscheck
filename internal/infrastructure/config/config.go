package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	WorkerCount int
	QueueSize   int

	DatabaseURL string

	RedisURL      string
	RedisPassword string

	GeminiAPIKey string
	GeminiModel  string
	OllamaURL    string
	OllamaModel  string
	Provider     string // "gemini" | "ollama" | "chained"

	JWTSecret string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	return Config{
		Port:        getEnv("PORT", "8080"),
		WorkerCount: getEnvInt("WORKER_COUNT", 10),
		QueueSize:   getEnvInt("QUEUE_SIZE", 500),

		DatabaseURL: mustEnv("DATABASE_URL"),

		RedisURL:      mustEnv("REDIS_URL"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
		GeminiModel:  getEnv("GEMINI_MODEL", "gemini-1.5-flash"),
		OllamaURL:    getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:  getEnv("OLLAMA_MODEL", "llama3.2"),
		Provider:     getEnv("PROVIDER", "chained"),

		JWTSecret: mustEnv("JWT_SECRET"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %q is not set", key)
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("environment variable %q must be an integer, got %q", key, v)
	}
	return n
}
