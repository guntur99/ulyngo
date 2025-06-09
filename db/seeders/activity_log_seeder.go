package seeders

import (
	"encoding/json"
	"log"
	"time"
	"ulyngo/models" // Ganti dengan nama modul Anda

	"gorm.io/gorm"
)

// SeedUserActivityLogs mengisi data log aktivitas pengguna awal ke database.
func SeedUserActivityLogs(db *gorm.DB) {
	log.Println("Seeding user activity logs...")

	// Ambil beberapa user yang sudah di-seed
	var users []models.User
	if err := db.Limit(2).Find(&users).Error; err != nil {
		log.Printf("Failed to fetch users for activity logs: %v", err)
		return
	}
	if len(users) == 0 {
		log.Println("No users found to create activity logs. Skipping activity log seeding.")
		return
	}

	// Ambil beberapa marker yang sudah di-seed (jika ada, setelah Marker seeder)
	var markers []models.Marker
	if err := db.Limit(2).Find(&markers).Error; err != nil {
		log.Printf("Failed to fetch markers for activity logs: %v", err)
		// Lanjutkan seeding tanpa markers jika gagal
	}

	activityLogs := []models.UserActivityLog{}

	// Log aktivitas untuk user pertama
	if len(users) > 0 {
		user1ID := users[0].ID
		// Contoh: aktivitas login sudah ditangani di user_seeder.go
		// activityLogs = append(activityLogs, models.UserActivityLog{
		// 	UserID:       user1ID,
		// 	ActivityType: "login",
		// 	ActivityData: json.RawMessage(`{"ip_address": "192.168.1.1", "device": "web"}`),
		// 	Timestamp:    time.Now().Add(-2 * time.Hour),
		// })

		if len(markers) > 0 {
			activityLogs = append(activityLogs, models.UserActivityLog{
				UserID:       user1ID,
				ActivityType: "view_marker",
				TargetID:     &markers[0].ID, // Menggunakan pointer ke ID marker pertama
				ActivityData: json.RawMessage(`{"marker_name": "` + markers[0].Name + `"}`),
				Timestamp:    time.Now().Add(-1 * time.Hour),
			})
		}
		activityLogs = append(activityLogs, models.UserActivityLog{
			UserID:       user1ID,
			ActivityType: "search_route",
			ActivityData: json.RawMessage(`{"query": "dari bandung ke jakarta", "result_count": 3}`),
			Timestamp:    time.Now().Add(-30 * time.Minute),
		})
	}

	// Log aktivitas untuk user kedua (jika ada)
	if len(users) > 1 {
		user2ID := users[1].ID
		// Contoh: aktivitas login sudah ditangani di user_seeder.go
		// activityLogs = append(activityLogs, models.UserActivityLog{
		// 	UserID:       user2ID,
		// 	ActivityType: "login",
		// 	ActivityData: json.RawMessage(`{"ip_address": "10.0.0.5", "device": "mobile"}`),
		// 	Timestamp:    time.Now().Add(-4 * time.Hour),
		// })
		activityLogs = append(activityLogs, models.UserActivityLog{
			UserID:       user2ID,
			ActivityType: "view_marker",
			TargetID:     nil, // Contoh aktivitas tanpa target_id
			ActivityData: json.RawMessage(`{"marker_name": "unknown_marker_viewed", "from_ai_recommendation": true}`),
			Timestamp:    time.Now().Add(-2 * time.Hour),
		})
	}

	for _, logEntry := range activityLogs {
		// Untuk UserActivityLog, kita tidak perlu cek duplikasi berdasarkan nama,
		// karena setiap log adalah event unik berdasarkan timestamp.
		// Cukup buat record baru.
		if err := db.Create(&logEntry).Error; err != nil {
			log.Printf("Failed to seed activity log for user %s, type %s: %v", logEntry.UserID, logEntry.ActivityType, err)
		} else {
			log.Printf("Seeded activity log for user %s, type %s", logEntry.UserID, logEntry.ActivityType)
		}
	}
	log.Println("User activity log seeding completed.")
}
