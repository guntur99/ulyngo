package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MarkerCategory mendefinisikan kategori untuk pengelompokan marker.
type MarkerCategory struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"` // ID kategori (UUID)
	Name        string         `gorm:"type:varchar(100);not null;unique" json:"name"`            // Nama kategori, unik dan tidak null
	Description *string        `json:"description"`                                              // Deskripsi kategori, bisa null
	CreatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`     // Waktu pembuatan record
	UpdatedAt   time.Time      `json:"updated_at"`                                               // Waktu pembaruan record
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`                        // Untuk soft delete

	// Relasi (opsional untuk GORM, digunakan untuk memuat marker dalam kategori ini)
	Markers []Marker `gorm:"foreignKey:CategoryID" json:"markers,omitempty"`
}

// BeforeCreate hook untuk MarkerCategory: Otomatis menghasilkan UUID untuk MarkerCategory.ID jika belum ada.
func (mc *MarkerCategory) BeforeCreate(tx *gorm.DB) (err error) {
	if mc.ID == uuid.Nil {
		mc.ID = uuid.New()
	}
	return
}
