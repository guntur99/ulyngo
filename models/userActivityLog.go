package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserActivityLog menyimpan log aktivitas pengguna yang terperinci.
type UserActivityLog struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"` // ID log aktivitas (UUID)
	UserID       uuid.UUID       `gorm:"type:uuid;not null" json:"user_id"`                        // ID pengguna terkait, tidak null
	ActivityType string          `gorm:"type:varchar(100);not null" json:"activity_type"`          // Tipe aktivitas (misal: 'view_marker')
	TargetID     *uuid.UUID      `gorm:"type:uuid" json:"target_id"`                               // ID entitas target (marker/route), bisa null
	ActivityData json.RawMessage `gorm:"type:jsonb" json:"activity_data"`                          // Data aktivitas tambahan (JSONB)
	Timestamp    time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP" json:"timestamp"`      // Waktu aktivitas

	// Relasi (opsional)
	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	// Target ini bersifat polimorfik (bisa Marker atau Route), sehingga relasi GORM tidak didefinisikan langsung di sini.
	// Penanganan target biasanya dilakukan secara manual dalam kode aplikasi.
}

// BeforeCreate hook untuk UserActivityLog: Otomatis menghasilkan UUID untuk UserActivityLog.ID jika belum ada.
func (ual *UserActivityLog) BeforeCreate(tx *gorm.DB) (err error) {
	if ual.ID == uuid.Nil {
		ual.ID = uuid.New()
	}
	return
}
