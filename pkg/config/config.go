package config

import "os"

type Config struct {
	Port                    string
	Env                     string
	FirebaseCredentialsPath string
	PostgresUrl             string
	MongoURI                string
	MetricsPort             string
}

func Load() *Config {
	return &Config{
		Port:                      getEnv("PORT", "8080"),
		Env:                       getEnv("ENV", "development"),
		FirebaseCredentialsPath:   getEnv("FIREBASE_CREDENTIALS_PATH", ""),
		PostgresUrl:               getEnv("POSTGRES_URL", "http://localhost:5432"),
		MongoURI:                  getEnv("MONGO_URI", ""),
		MetricsPort:               getEnv("METRICS_PORT", "9090"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}