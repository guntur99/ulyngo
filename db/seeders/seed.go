package seeders

import (
	"log"

	"gorm.io/gorm"
)

// RunAllSeeders menjalankan semua fungsi seeder.
func RunAllSeeders(db *gorm.DB) {
	log.Println("Starting database seeding...")

	SeedUsers(db)
	SeedMarkerCategories(db)
	SeedMarkerTags(db)
	// Tambahkan panggilan ke fungsi seeder lain di sini jika Anda membuatnya

	log.Println("All database seeding completed.")
}
