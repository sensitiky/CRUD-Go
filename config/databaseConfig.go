package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func DatabaseConnection() (*sql.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	log.Printf("Database configuration: Host=%s, Port=%s, User=%s, DBName=%s", dbHost, dbPort, dbUser, dbName)

	log.Println("Attempting to connect to the database...")
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName))
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	log.Println("Database connection established")

	log.Println("Pinging the database...")
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping the database:", err)
	}
	log.Println("Database ping successful")
	return db, nil
}
