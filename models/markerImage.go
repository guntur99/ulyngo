package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MarkerImage menyimpan URL dan deskripsi gambar yang terkait dengan sebuah marker.
type MarkerImage struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"` // ID gambar marker (UUID)
	MarkerID    uuid.UUID      `gorm:"type:uuid;not null" json:"marker_id"`                      // ID marker terkait, tidak null
	ImageURL    string         `gorm:"type:varchar(255);not null" json:"image_url"`              // URL gambar, tidak null
	Description *string        `json:"description"`                                              // Deskripsi gambar, bisa null
	UploadedAt  time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"uploaded_at"`    // Waktu upload gambar
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`                        // Untuk soft delete
}

// BeforeCreate hook untuk MarkerImage: Otomatis menghasilkan UUID untuk MarkerImage.ID jika belum ada.
func (mi *MarkerImage) BeforeCreate(tx *gorm.DB) (err error) {
	if mi.ID == uuid.Nil {
		mi.ID = uuid.New()
	}
	return
}
