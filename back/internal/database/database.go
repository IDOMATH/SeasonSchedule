package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// Config holds database connection configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// LoadConfig loads database configuration from environment variables with defaults
func LoadConfig() Config {
	return Config{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5432"),
		User:     getEnvOrDefault("DB_USER", "seasonschedule"),
		Password: getEnvOrDefault("DB_PASSWORD", "seasonschedule_pass"),
		DBName:   getEnvOrDefault("DB_NAME", "seasonschedule_db"),
	}
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Connect establishes a connection to the PostgreSQL database using the provided config
func Connect(config Config) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Verify the connection is actually working
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Printf("Database connection established successfully to %s:%s/%s", config.Host, config.Port, config.DBName)
	return db, nil
}
