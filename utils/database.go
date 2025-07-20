package utils

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres" // Ganti dengan driver database Anda (misal: mysql, sqlite)
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDatabase menginisialisasi koneksi database GORM
func ConnectDatabase() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{}) // Gunakan driver yang sesuai
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Database connected!")
}
