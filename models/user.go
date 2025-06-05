package models

import (
	"time"

	"github.com/google/uuid" // Import library UUID
	"gorm.io/gorm"
)

// User merepresentasikan entitas pengguna dalam database.
type User struct {
	ID           string `gorm:"type:uuid;primaryKey"` // Mengubah ID menjadi string (UUID)
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Username     string         `gorm:"unique;not null"`
	PasswordHash string         `gorm:"not null"`
	Role         string         `gorm:"not null;default:'user'"` // Role pengguna (misal: user, admin)
}

// BeforeCreate hook akan otomatis menghasilkan UUID untuk User.ID
// sebelum record disimpan ke database, jika ID belum diset.
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" { // Hanya generate jika ID belum diset
		u.ID = uuid.New().String()
	}
	return
}
