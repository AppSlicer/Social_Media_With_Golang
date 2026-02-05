package config

import (
	"context"
	"fmt"
	"log"
	"os" // Added this import
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB holds the database connections
type DB struct {
	Postgres *gorm.DB
	Mongo    *mongo.Client
}

// InitDB initializes and returns the database connections
func InitDB() (*DB, error) {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, assuming environment variables are set.")
	}

	pgConnStr := os.Getenv("POSTGRES_CONN_STR")
	if pgConnStr == "" {
		return nil, fmt.Errorf("POSTGRES_CONN_STR environment variable not set")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		return nil, fmt.Errorf("MONGO_URI environment variable not set")
	}

	postgresDB, err := initPostgres(pgConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	mongoClient, err := initMongo(mongoURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	return &DB{
		Postgres: postgresDB,
		Mongo:    mongoClient,
	}, nil
}

// initPostgres initializes the PostgreSQL database connection using GORM
func initPostgres(connStr string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Ping the database to verify connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err = sqlDB.Ping(); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to PostgreSQL!")
	return db, nil
}

// initMongo initializes the MongoDB connection
func initMongo(uri string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the primary to verify connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to MongoDB!")
	return client, nil
}

// CloseDB closes the database connections
func (db *DB) CloseDB() {
	if db.Postgres != nil {
		sqlDB, err := db.Postgres.DB()
		if err != nil {
			log.Printf("Error getting SQL DB from GORM: %v\n", err)
		} else {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Error closing PostgreSQL connection: %v\n", err)
			} else {
				log.Println("PostgreSQL connection closed.")
			}
		}
	}

	if db.Mongo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.Mongo.Disconnect(ctx); err != nil {
			log.Printf("Error closing MongoDB connection: %v\n", err)
		} else {
			log.Println("MongoDB connection closed.")
		}
	}
}
