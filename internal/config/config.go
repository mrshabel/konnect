package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DbName             string
	DbPassword         string
	DbUsername         string
	DbPort             string
	DbHost             string
	Port               int
	JWTSecret          string
	JWTExpiryMinutes   time.Duration
	MaxNearbyRadius    float64
	GoogleClientID     string
	GoogleClientSecret string
	GoogleCallbackURL  string
	CloudinaryURL      string
	RedisAddr          string
	RedisURL           string
	CourierAPIKey      string
}

// New returns a config object from the env and a non-nil error if the env value is not present
func New() (*Config, error) {
	_ = godotenv.Load()

	// database configs
	dbName := getEnv("DB_NAME", "")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbUsername := getEnv("DB_USERNAME", "")
	dbPort := getEnv("DB_PORT", "")
	dbHost := getEnv("DB_HOST", "")

	redisHost := getEnv("REDIS_HOST", "")
	redisPort := getEnvInt("REDIS_PORT", 6379)

	// server configs
	port := getEnvInt("PORT", 8000)
	jwtSecret := getEnv("JWT_SECRET", "")
	jwtExpiry := getEnvInt("JWT_EXPIRY_MINUTES", 60)
	maxRadius := getEnvFloat("MAX_RADIUS_METERS", 5000)

	// oauth configs
	googleClientID := getEnv("GOOGLE_CLIENT_ID", "")
	googleClientSecret := getEnv("GOOGLE_CLIENT_SECRET", "")
	googleCallbackURL := getEnv("GOOGLE_CALLBACK_URL", "")

	cloudinaryURL := getEnv("CLOUDINARY_URL", "")

	// notifs
	courierAPIKey := getEnv("COURIER_API_KEY", "")

	return &Config{
		DbName:             dbName,
		DbPassword:         dbPassword,
		DbUsername:         dbUsername,
		DbPort:             dbPort,
		DbHost:             dbHost,
		Port:               port,
		JWTSecret:          jwtSecret,
		JWTExpiryMinutes:   time.Duration(jwtExpiry) * time.Minute,
		MaxNearbyRadius:    maxRadius,
		GoogleClientID:     googleClientID,
		GoogleClientSecret: googleClientSecret,
		GoogleCallbackURL:  googleCallbackURL,
		CloudinaryURL:      cloudinaryURL,
		RedisAddr:          fmt.Sprintf("%s:%d", redisHost, redisPort),
		RedisURL:           fmt.Sprintf("redis://%s:%d", redisHost, redisPort),
		CourierAPIKey:      courierAPIKey,
	}, nil
}

func getEnv(key string, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		if fallback == "" {
			panic("env variable not set: " + key)
		}
		return fallback
	}
	return val
}

func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		if fallback == 0 {
			panic("env variable not set: " + key)
		}
		return fallback
	}
	intVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fallback
	}
	return int(intVal)
}

func getEnvFloat(key string, fallback float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		if fallback == 0 {
			panic("env variable not set: " + key)
		}
		return fallback
	}
	floatVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return fallback
	}
	return floatVal
}

func getEnvArr(key string, fallback []string) []string {
	val := os.Getenv(key)
	if val == "" {
		if len(fallback) == 0 {
			panic("env variable not set: " + key)
		}
		return fallback
	}
	return strings.Split(val, ",")
}
