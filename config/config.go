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

	// Go Service API
	Port           string
	RequestTimeout int

	// MQTT
	MQTTBroker   string
	MQTTClientID string

	// RabbitMQ
	RabbitMQURL string

	// Geofence
	GeofenceLat    float64
	GeofenceLng    float64
	GeofenceRadius float64 // meters
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

		Port:           getEnv("PORT", "8080"),
		RequestTimeout: getEnvInt("REQUEST_TIMEOUT", 30), // Timeout in seconds

		MQTTBroker:   getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		MQTTClientID: getEnv("MQTT_CLIENT_ID", "fleet-backend"),

		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),

		GeofenceLat:    getEnvFloat("GEOFENCE_LAT", -6.2088),
		GeofenceLng:    getEnvFloat("GEOFENCE_LNG", 106.8456),
		GeofenceRadius: getEnvFloat("GEOFENCE_RADIUS_METER", 50),
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

func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return i
}
