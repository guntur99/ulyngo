package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Marker merepresentasikan entitas marker/tempat di peta.
type Place struct {
	ID                    string `gorm:"type:uuid;primaryKey" json:"id"` // ID marker (UUID)
	DitambahkanPada       time.Time
	DiperbaruiPada        time.Time
	DeletedAt             gorm.DeletedAt `gorm:"index"`
	Nama                  string         `json:"nama"`
	Deskripsi             string         `json:"deskripsi"`
	Latitude              float64        `json:"latitude"`
	Longitude             float64        `json:"longitude"`
	DitambahkanOlehUserId string         `json:"ditambahkan_oleh_user_id" gorm:"type:uuid;not null"` // ID pengguna yang menambahkan marker
}

// BeforeCreate hook untuk Marker (opsional, jika Anda ingin ID marker juga UUID)
// Jika Anda ingin GORM secara otomatis membuat ID uint default, Anda bisa menghapus ini.
// Jika Anda ingin ID UUID, pastikan untuk mengimpor "github.com/google/uuid"
func (m *Place) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return
}
