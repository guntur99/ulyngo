package models

import (
	"time"

	"gorm.io/gorm"
)

// Marker merepresentasikan entitas marker/tempat di peta.
type Marker struct {
	ID            string `gorm:"type:uuid;primaryKey"` // ID marker (UUID)
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Latitude      float64        `json:"latitude"`
	Longitude     float64        `json:"longitude"`
	AddedByUserID string         `json:"added_by_user_id" gorm:"type:uuid;not null"` // ID pengguna yang menambahkan marker
}

// BeforeCreate hook untuk Marker (opsional, jika Anda ingin ID marker juga UUID)
// Jika Anda ingin GORM secara otomatis membuat ID uint default, Anda bisa menghapus ini.
// Jika Anda ingin ID UUID, pastikan untuk mengimpor "github.com/google/uuid"
// func (m *Marker) BeforeCreate(tx *gorm.DB) (err error) {
//     if m.ID == "" {
//         m.ID = uuid.New().String()
//     }
//     return
// }
