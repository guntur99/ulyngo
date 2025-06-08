package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MarkerReview menyimpan ulasan dan rating yang diberikan pengguna untuk sebuah marker.
type MarkerReview struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`               // ID ulasan (UUID)
	MarkerID  uuid.UUID      `gorm:"type:uuid;not null" json:"marker_id"`                                    // ID marker terkait, tidak null
	UserID    uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`                                      // ID pengguna yang memberikan ulasan, tidak null
	Rating    int            `gorm:"type:smallint;not null;check:rating >= 1 AND rating <= 5" json:"rating"` // Rating 1-5 bintang
	Comment   *string        `json:"comment"`                                                                // Komentar ulasan, bisa null
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`                   // Waktu pembuatan record
	UpdatedAt time.Time      `json:"updated_at"`                                                             // Waktu pembaruan record
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`                                      // Untuk soft delete

	// Relasi
	Marker Marker `gorm:"foreignKey:MarkerID" json:"marker,omitempty"`
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// BeforeCreate hook untuk MarkerReview: Otomatis menghasilkan UUID untuk MarkerReview.ID jika belum ada.
func (mr *MarkerReview) BeforeCreate(tx *gorm.DB) (err error) {
	if mr.ID == uuid.Nil {
		mr.ID = uuid.New()
	}
	return
}
