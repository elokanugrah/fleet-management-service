package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// PostgreSQL
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	// MQTT
	MQTTBroker   string
	MQTTClientID string

	// RabbitMQ
	RabbitMQURL string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "fleet_db"),
		DBUser:     getEnv("DB_USER", "fleet_user"),
		DBPassword: getEnv("DB_PASSWORD", "fleet_pass"),

		MQTTBroker:   getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		MQTTClientID: getEnv("MQTT_CLIENT_ID", "fleet-backend"),

		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

func (c *Config) DBConnectionString() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=disable"
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return fallback
	}
	return f
}
