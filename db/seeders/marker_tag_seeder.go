package seeders

import (
	"log"
	"ulyngo/models" // Ganti dengan nama modul Anda

	"gorm.io/gorm"
)

// SeedMarkerTags mengisi data tag marker awal ke database.
func SeedMarkerTags(db *gorm.DB) {
	tags := []models.MarkerTag{
		{Name: "Family-Friendly"},
		{Name: "Pet-Friendly"},
		{Name: "Sunset-View"},
		{Name: "Live Music"},
		{Name: "Hiking"},
		{Name: "Snorkeling"},
		{Name: "Budget-Friendly"},
		{Name: "Luxury"},
	}

	log.Println("Seeding marker tags...")
	for _, tag := range tags {
		var existingTag models.MarkerTag
		if err := db.Where("name = ?", tag.Name).First(&existingTag).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&tag).Error; err != nil {
					log.Printf("Failed to seed tag %s: %v", tag.Name, err)
				} else {
					log.Printf("Seeded tag: %s", tag.Name)
				}
			} else {
				log.Printf("Error checking tag %s: %v", tag.Name, err)
			}
		} else {
			log.Printf("Tag %s already exists, skipping.", tag.Name)
		}
	}
	log.Println("Marker tag seeding completed.")
}
