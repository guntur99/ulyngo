package models

import (
	"time"

	"github.com/google/uuid" // Import library UUID
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"` // ID pengguna (UUID)
	Username     string         `gorm:"unique;not null" json:"username"`                          // Nama pengguna, harus unik dan tidak null
	Email        string         `gorm:"unique;not null" json:"email"`                             // Alamat email, harus unik dan tidak null
	Whatsapp     *string        `gorm:"unique" json:"whatsapp"`                                   // Nomor WhatsApp, bisa null dan unik
	Password     string         `gorm:"not null" json:"-"`                                        // Hash kata sandi, tidak null, tidak disertakan dalam JSON
	Role         string         `gorm:"not null;default:'user'" json:"role"`                      // Peran pengguna (misal: 'user', 'admin'), default 'user'
	LastActiveAt *time.Time     `json:"last_active_at"`                                           // Waktu terakhir aktif, bisa null
	CreatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`     // Waktu pembuatan record
	UpdatedAt    time.Time      `json:"updated_at"`                                               // Waktu pembaruan record
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`                        // Untuk soft delete, indeks untuk pencarian cepat

	// Relasi: Digunakan oleh GORM untuk memuat relasi jika dikonfigurasi (Preload, Join).
	// "omitempty" memastikan tidak disertakan dalam JSON jika kosong.
	Preferences   []Preference      `gorm:"foreignKey:UserID" json:"preferences,omitempty"`
	Markers       []Marker          `gorm:"foreignKey:AddedByUserID" json:"markers,omitempty"`
	MarkerReviews []MarkerReview    `gorm:"foreignKey:UserID" json:"marker_reviews,omitempty"`
	Routes        []Route           `gorm:"foreignKey:UserID" json:"routes,omitempty"`
	ActivityLogs  []UserActivityLog `gorm:"foreignKey:UserID" json:"activity_logs,omitempty"`
}

// BeforeCreate hook untuk User: Otomatis menghasilkan UUID untuk User.ID jika belum ada.
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil { // Jika ID belum diset (misal: default UUID di DB tidak digunakan)
		u.ID = uuid.New() // Hasilkan UUID baru
	}
	return
}
