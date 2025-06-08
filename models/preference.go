package models

import (
	"time"

	"github.com/google/uuid" // Import library UUID
	"gorm.io/gorm"
)

// Preference menyimpan preferensi spesifik pengguna dalam format key-value.
type Preference struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`        // ID preferensi (UUID)
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index:idx_user_key,unique" json:"user_id"`     // ID pengguna, bagian dari indeks unik komposit
	Key       string         `gorm:"type:varchar(100);not null;index:idx_user_key,unique" json:"key"` // Kunci preferensi, bagian dari indeks unik komposit
	Value     string         `gorm:"type:text;not null" json:"value"`                                 // Nilai preferensi (disimpan sebagai TEXT)
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`            // Waktu pembuatan record
	UpdatedAt time.Time      `json:"updated_at"`                                                      // Waktu pembaruan record
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`                               // Untuk soft delete
}

// BeforeCreate hook untuk Preference: Otomatis menghasilkan UUID untuk Preference.ID jika belum ada.
func (p *Preference) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}
