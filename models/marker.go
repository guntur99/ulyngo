package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Marker merepresentasikan entitas lokasi atau tempat menarik di peta.
type Marker struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"` // ID marker (UUID)
	Name        string    `gorm:"not null" json:"name"`                                     // Nama marker, tidak null
	Description *string   `json:"description"`                                              // Deskripsi marker, bisa null
	Latitude    float64   `gorm:"not null" json:"latitude"`                                 // Koordinat lintang, tidak null
	Longitude   float64   `gorm:"not null" json:"longitude"`                                // Koordinat bujur, tidak null
	// Geometry (GEOGRAPHY): GORM tidak memiliki tipe Go langsung. Penanganan untuk PostGIS
	// biasanya dilakukan dengan plugin GORM atau menggunakan raw SQL untuk kolom geometry.
	CategoryID    uuid.UUID      `gorm:"type:uuid;not null" json:"category_id"`                // ID kategori marker, tidak null
	AvgRating     float64        `gorm:"type:numeric(2,1);default:0.0" json:"avg_rating"`      // Rata-rata rating, default 0.0
	TotalReviews  int            `gorm:"type:integer;default:0" json:"total_reviews"`          // Total ulasan, default 0
	ViewCount     int64          `gorm:"type:bigint;default:0" json:"view_count"`              // Jumlah tampilan, default 0
	AddedByUserID uuid.UUID      `gorm:"type:uuid;not null" json:"added_by_user_id"`           // ID pengguna yang menambahkan, tidak null
	CreatedAt     time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"` // Waktu pembuatan record
	UpdatedAt     time.Time      `json:"updated_at"`                                           // Waktu pembaruan record
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`                    // Untuk soft delete

	// Relasi
	Category MarkerCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Images   []MarkerImage  `gorm:"foreignKey:MarkerID" json:"images,omitempty"`
	Reviews  []MarkerReview `gorm:"foreignKey:MarkerID" json:"reviews,omitempty"`
	Tags     []MarkerTag    `gorm:"many2many:marker_has_tags;" json:"tags,omitempty"` // Many-to-many relationship
}

// BeforeCreate hook untuk Marker: Otomatis menghasilkan UUID untuk Marker.ID jika belum ada.
func (m *Marker) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}
