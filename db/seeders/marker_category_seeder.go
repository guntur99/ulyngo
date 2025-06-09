package seeders

import (
	"log"
	"time"
	"ulyngo/models" // Ganti dengan nama modul Anda

	"gorm.io/gorm"
)

// SeedMarkerCategories mengisi data kategori marker awal ke database.
func SeedMarkerCategories(db *gorm.DB) {
	categories := []models.MarkerCategory{
		{
			Name:        "Restaurant",
			Description: strPtr("Tempat makan dan minum."),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "Wisata Alam",
			Description: strPtr("Destinasi alam seperti pantai, gunung, hutan."),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "Sejarah & Budaya",
			Description: strPtr("Tempat bersejarah, museum, situs budaya."),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "Akomodasi",
			Description: strPtr("Hotel, penginapan, villa."),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "Transportasi",
			Description: strPtr("Stasiun, bandara, terminal."),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "Hiburan",
			Description: strPtr("Pusat perbelanjaan, bioskop, taman hiburan."),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	log.Println("Seeding marker categories...")
	for _, category := range categories {
		var existingCategory models.MarkerCategory
		if err := db.Where("name = ?", category.Name).First(&existingCategory).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&category).Error; err != nil {
					log.Printf("Failed to seed category %s: %v", category.Name, err)
				} else {
					log.Printf("Seeded category: %s", category.Name)
				}
			} else {
				log.Printf("Error checking category %s: %v", category.Name, err)
			}
		} else {
			log.Printf("Category %s already exists, skipping.", category.Name)
		}
	}
	log.Println("Marker category seeding completed.")
}

// strPtr returns a pointer to the given string.
func strPtr(s string) *string {
	return &s
}
