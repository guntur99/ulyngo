package seeders

import (
	"log"
	"ulyngo/models" // Ganti dengan nama modul Anda

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedUsers mengisi data pengguna awal ke database.
func SeedUsers(db *gorm.DB) {
	// Hash password untuk pengguna admin
	hashedPasswordAdmin, err := bcrypt.GenerateFromPassword([]byte("adminpassword"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash admin password: %v", err)
	}

	// Hash password untuk pengguna biasa
	hashedPasswordUser, err := bcrypt.GenerateFromPassword([]byte("userpassword"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash user password: %v", err)
	}

	users := []models.User{
		{
			Username: "superadmin",
			Email:    "superadmin@ulyn.com",
			Whatsapp: nil, // Bisa null
			Password: string(hashedPasswordAdmin),
			Role:     "admin",
		},
		{
			Username: "raffa",
			Email:    "raffa@ulyn.com",
			Whatsapp: strPtr("6281220544440"), // Contoh WhatsApp
			Password: string(hashedPasswordUser),
			Role:     "user",
		},
	}

	log.Println("Seeding users...")
	for _, user := range users {
		// Cek apakah pengguna sudah ada berdasarkan username
		var existingUser models.User
		if err := db.Where("username = ?", user.Username).First(&existingUser).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Pengguna belum ada, buat yang baru
				if err := db.Create(&user).Error; err != nil {
					log.Printf("Failed to seed user %s: %v", user.Username, err)
				} else {
					log.Printf("Seeded user: %s", user.Username)
				}
			} else {
				log.Printf("Error checking user %s: %v", user.Username, err)
			}
		} else {
			log.Printf("User %s already exists, skipping.", user.Username)
		}
	}
	log.Println("User seeding completed.")
}
