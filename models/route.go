package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Route menyimpan rute perjalanan yang dibuat dan disimpan oleh pengguna.
type Route struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"` // ID rute (UUID)
	Name            string    `gorm:"type:varchar(255);not null" json:"name"`                   // Nama rute, tidak null
	OriginText      string    `gorm:"type:varchar(255);not null" json:"origin_text"`            // Teks asal, tidak null
	DestinationText string    `gorm:"type:varchar(255);not null" json:"destination_text"`       // Teks tujuan, tidak null
	OriginLat       float64   `gorm:"type:double precision" json:"origin_lat"`                  // Lintang asal
	OriginLng       float64   `gorm:"type:double precision" json:"origin_lng"`                  // Bujur asal
	DestinationLat  float64   `gorm:"type:double precision" json:"destination_lat"`             // Lintang tujuan
	DestinationLng  float64   `gorm:"type:double precision" json:"destination_lng"`             // Bujur tujuan
	// RouteGeometry (GEOGRAPHY): GORM tidak memiliki tipe Go langsung.
	RouteDataJSON   json.RawMessage `gorm:"type:jsonb" json:"route_data_json"`                    // Data rute lengkap (JSONB)
	DistanceMeters  int64           `gorm:"type:bigint" json:"distance_meters"`                   // Jarak total dalam meter
	DurationSeconds int64           `gorm:"type:bigint" json:"duration_seconds"`                  // Durasi total dalam detik
	UserID          uuid.UUID       `gorm:"type:uuid;not null" json:"user_id"`                    // ID pengguna yang menyimpan rute, tidak null
	IsPublic        bool            `gorm:"not null;default:false" json:"is_public"`              // Apakah rute publik, default false
	CreatedAt       time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"` // Waktu pembuatan record
	UpdatedAt       time.Time       `json:"updated_at"`                                           // Waktu pembaruan record
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`                    // Untuk soft delete

	// Relasi
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// BeforeCreate hook untuk Route: Otomatis menghasilkan UUID untuk Route.ID jika belum ada.
func (r *Route) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}
