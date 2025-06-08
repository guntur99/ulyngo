package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MarkerTag mendefinisikan tag yang dapat dikaitkan dengan marker.
type MarkerTag struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"` // ID tag (UUID)
	Name      string         `gorm:"type:varchar(100);not null;unique" json:"name"`            // Nama tag, unik dan tidak null
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`     // Waktu pembuatan record
	UpdatedAt time.Time      `json:"updated_at"`                                               // Waktu pembaruan record
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`                        // Untuk soft delete

	// Relasi Many-to-Many
	Markers []Marker `gorm:"many2many:marker_has_tags;" json:"markers,omitempty"`
}

// BeforeCreate hook untuk MarkerTag: Otomatis menghasilkan UUID untuk MarkerTag.ID jika belum ada.
func (mt *MarkerTag) BeforeCreate(tx *gorm.DB) (err error) {
	if mt.ID == uuid.Nil {
		mt.ID = uuid.New()
	}
	return
}
