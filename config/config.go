package config

import (
	"log"
	"os"
)

var Env = LoadConfig()

type Config struct {
	DatabaseURI          string
	DatabaseName         string
	ServerPort           string
	DefaultAdminUsername string
	DefaultAdminPassword string
}

func LoadConfig() *Config {
	config := &Config{
		DatabaseURI:          getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName:         getEnv("MONGO_DB_NAME", "mongodb"),
		ServerPort:           getEnv("SERVER_PORT", "8080"),
		DefaultAdminUsername: getEnv("DEFAULT_ADMIN_USERNAME", "admin"),
		DefaultAdminPassword: getEnv("DEFAULT_ADMIN_PASSWORD", "password"),
	}

	// Log loaded configuration (excluding sensitive information in production)
	log.Printf("Config loaded: %+v\n", config)
	return config
}

func getEnv(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
