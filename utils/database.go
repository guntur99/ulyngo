package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB   *gorm.DB
	once sync.Once
)

// ConnectDatabase menginisialisasi koneksi database GORM
func ConnectDatabase() {
	once.Do(func() {
		dsn := os.Getenv("DATABASE_URL")
		log.Println("Using DATABASE_URL:", dsn)
		if dsn == "" {
			log.Fatal("DATABASE_URL is not set in environment variables")
		}

		// Railway biasanya kasih "postgres://", gorm butuh "postgresql://"
		if strings.HasPrefix(dsn, "postgres://") {
			dsn = strings.Replace(dsn, "postgres://", "postgresql://", 1)
		}

		var err error
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("❌ Failed to connect to database: %v", err)
		}

		// Cek koneksi
		sqlDB, err := DB.DB()
		if err != nil {
			log.Fatalf("❌ Failed to get db instance: %v", err)
		}

		if err := sqlDB.Ping(); err != nil {
			log.Fatalf("❌ Database ping failed: %v", err)
		}

		fmt.Println("✅ Database connected successfully!")
	})
}
