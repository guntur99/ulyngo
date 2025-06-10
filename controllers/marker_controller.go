package controllers

import (
	"net/http"
	"time" // Import time untuk UpdateMarker

	"ulyngo/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm" // Import gorm untuk akses database
)

// MarkerController struct akan menampung dependensi database
type MarkerController struct {
	DB *gorm.DB
}

// NewMarkerController adalah konstruktor untuk MarkerController.
// Menerima instance GORM DB untuk dependency injection.
func NewMarkerController(db *gorm.DB) *MarkerController {
	return &MarkerController{DB: db}
}

// DirectionsRequest adalah struktur untuk data permintaan rute.
type DirectionsRequest struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
}

// GetMarkers adalah metode dari MarkerController yang mengambil semua marker dari database.
func (tc *MarkerController) GetMarkers(c *gin.Context) {
	var markers []models.Marker
	// Menggunakan dependensi DB yang di-inject untuk mengambil markers
	if err := tc.DB.Find(&markers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch markers: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, markers)
}

// AddMarkerInput adalah struktur untuk data yang diterima saat menambah marker baru.
type AddMarkerInput struct {
	Name          string    `json:"name" binding:"required"`
	Description   string    `json:"description" binding:"required"`
	Latitude      float64   `json:"latitude" binding:"required"`
	Longitude     float64   `json:"longitude" binding:"required"`
	CategoryID    uuid.UUID `json:"category_id"`      // Tambahkan CategoryID
	AddedByUserId string    `json:"added_by_user_id"` // Ini akan diisi otomatis dari token JWT

}

// AddMarker adalah metode dari MarkerController yang menambahkan marker baru ke database.
// Membutuhkan token JWT dan akan menyimpan DitambahkanOlehUserId.
func (tc *MarkerController) AddMarker(c *gin.Context) {

	var input AddMarkerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mendapatkan userID dari konteks Gin yang diset oleh AuthMiddleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	addedByUserID := userID.(string) // Konversi ke string (UUID)

	// Konversi string ke uuid.UUID
	addedByUserUUID, err := uuid.Parse(addedByUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// c.JSON(http.StatusOK, gin.H{"message": input}) // Hapus ini, karena akan mengirim respons ganda
	marker := models.Marker{
		Name:          input.Name,
		Description:   &input.Description,
		Latitude:      input.Latitude,
		Longitude:     input.Longitude,
		AddedByUserID: addedByUserUUID,  // Mengisi DitambahkanOlehUserId
		CategoryID:    input.CategoryID, // Mengisi CategoryID
		CreatedAt:     time.Now(),       // Set waktu pembuatan
		UpdatedAt:     time.Now(),       // Set waktu pembaruan
	}
	// Menyimpan marker ke database
	if err := tc.DB.Create(&marker).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add marker: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Marker added successfully", "marker": marker})
}

// UpdateMarkerInput adalah struktur untuk data yang diterima saat memperbarui marker.
type UpdateMarkerInput struct {
	Name          *string    `json:"name"` // Gunakan pointer agar bisa null (opsional)
	Description   *string    `json:"description"`
	Latitude      *float64   `json:"latitude"`
	Longitude     *float64   `json:"longitude"`
	CategoryID    *uuid.UUID `json:"category_id"`              // Tambahkan CategoryID
	AddedByUserID *string    `json:"ditambahkan_oleh_user_id"` // Ini tidak perlu di-update, hanya untuk referensi
	UpdatedAt     *time.Time `json:"updated_at"`               // Ini tidak perlu di-update, hanya untuk referensi
}

// UpdateMarker adalah metode dari MarkerController yang memperbarui marker yang sudah ada.
// Membutuhkan token JWT dan hanya bisa memperbarui marker yang dimiliki pengguna.
func (tc *MarkerController) UpdateMarker(c *gin.Context) {
	markerID := c.Param("id") // Dapatkan ID marker dari parameter URL

	var input UpdateMarkerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	ownerUserID := userID.(string)

	var marker models.Marker
	// Cari marker berdasarkan ID dan pastikan DitambahkanOlehUserId cocok (kepemilikan)
	if err := tc.DB.Where("id = ? AND ditambahkan_oleh_user_id = ?", markerID, ownerUserID).First(&marker).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Marker not found or you don't have permission to update it"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find marker: " + err.Error()})
		}
		return
	}

	// Perbarui field yang disediakan dalam input
	if input.Name != nil {
		marker.Name = *input.Name
	}
	if input.Description != nil {
		marker.Description = input.Description
	}
	if input.Latitude != nil {
		marker.Latitude = *input.Latitude
	}
	if input.Longitude != nil {
		marker.Longitude = *input.Longitude
	}

	marker.UpdatedAt = time.Now() // Perbarui timestamp UpdatedAt

	if err := tc.DB.Save(&marker).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update marker: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Marker updated successfully", "marker": marker})
}

// DeleteMarker adalah metode dari MarkerController yang menghapus marker.
// Membutuhkan token JWT dan hanya bisa menghapus marker yang dimiliki pengguna.
func (tc *MarkerController) DeleteMarker(c *gin.Context) {
	markerID := c.Param("id")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	ownerUserID, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	var marker models.Marker
	// Jika role admin, izinkan hapus marker apapun; jika bukan, hanya marker miliknya
	role, _ := c.Get("role")
	var err error
	if role == "admin" {
		err = tc.DB.Where("id = ?", markerID).First(&marker).Error
	} else {
		err = tc.DB.Where("id = ? AND added_by_user_id = ?", markerID, ownerUserID).First(&marker).Error
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Marker not found or you don't have permission to delete it"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find marker: " + err.Error()})
		}
		return
	}

	if err := tc.DB.Delete(&marker).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete marker: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Marker deleted successfully"})
}
