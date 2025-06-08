package models

import (
	"time"

	"github.com/google/uuid"
)

// MarkerHasTag adalah model untuk tabel penghubung (junction table) antara Marker dan MarkerTag.
// Kolom MarkerID dan TagID membentuk primary key komposit.
type MarkerHasTag struct {
	MarkerID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"marker_id"`                // Foreign Key ke Marker, bagian dari PK komposit
	TagID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"tag_id"`                   // Foreign Key ke MarkerTag, bagian dari PK komposit
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"` // Waktu relasi dibuat
}
