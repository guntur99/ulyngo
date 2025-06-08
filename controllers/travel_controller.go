package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time" // Import time untuk UpdateMarker

	"ulyngo/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm" // Import gorm untuk akses database
)

// TravelController struct akan menampung dependensi database
type TravelController struct {
	DB *gorm.DB
}

// NewTravelController adalah konstruktor untuk TravelController.
// Menerima instance GORM DB untuk dependency injection.
func NewTravelController(db *gorm.DB) *TravelController {
	return &TravelController{DB: db}
}

// DirectionsRequest adalah struktur untuk data permintaan rute.
type DirectionsRequest struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
}

// GetDirections adalah metode dari TravelController yang mengambil rute dari Google Maps API.
func (tc *TravelController) GetDirections(c *gin.Context) {
	var req DirectionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	googleMapsAPIKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if googleMapsAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Google Maps API Key not set"})
		return
	}

	// Membangun URL untuk Google Directions API
	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/directions/json?origin=%s&destination=%s&key=%s",
		req.Origin, req.Destination, googleMapsAPIKey)

	// Melakukan permintaan HTTP ke Google Directions API
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch directions from Google Maps"})
		return
	}
	defer resp.Body.Close() // Pastikan body response ditutup

	// Membaca body response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read Google Maps response"})
		return
	}

	var result map[string]interface{}
	// Mem-parse JSON response
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Google Maps response"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetMarkers adalah metode dari TravelController yang mengambil semua marker dari database.
func (tc *TravelController) GetMarkers(c *gin.Context) {
	var markers []models.Place
	// Menggunakan dependensi DB yang di-inject untuk mengambil markers
	if err := tc.DB.Find(&markers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch markers: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, markers)
}

// AddMarkerInput adalah struktur untuk data yang diterima saat menambah marker baru.
type AddMarkerInput struct {
	Nama      string  `json:"nama" binding:"required"`
	Deskripsi string  `json:"deskripsi"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Category  string  `json:"category"` // Tambahkan Category
}

// AddMarker adalah metode dari TravelController yang menambahkan marker baru ke database.
// Membutuhkan token JWT dan akan menyimpan DitambahkanOlehUserId.
func (tc *TravelController) AddMarker(c *gin.Context) {

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

	// c.JSON(http.StatusOK, gin.H{"message": input}) // Hapus ini, karena akan mengirim respons ganda
	marker := models.Place{
		Nama:                  input.Nama,
		Deskripsi:             input.Deskripsi,
		Latitude:              input.Latitude,
		Longitude:             input.Longitude,
		DitambahkanOlehUserId: addedByUserID, // Mengisi DitambahkanOlehUserId
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
	Nama                  *string  `json:"nama"` // Gunakan pointer agar bisa null (opsional)
	Deskripsi             *string  `json:"deskripsi"`
	Latitude              *float64 `json:"latitude"`
	Longitude             *float64 `json:"longitude"`
	Category              *string  `json:"category"`                 // Tambahkan Category
	DitambahkanOlehUserId *string  `json:"ditambahkan_oleh_user_id"` // Ini tidak perlu di-update, hanya untuk referensi
}

// UpdateMarker adalah metode dari TravelController yang memperbarui marker yang sudah ada.
// Membutuhkan token JWT dan hanya bisa memperbarui marker yang dimiliki pengguna.
func (tc *TravelController) UpdateMarker(c *gin.Context) {
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

	var marker models.Place
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
	if input.Nama != nil {
		marker.Nama = *input.Nama
	}
	if input.Deskripsi != nil {
		marker.Deskripsi = *input.Deskripsi
	}
	if input.Latitude != nil {
		marker.Latitude = *input.Latitude
	}
	if input.Longitude != nil {
		marker.Longitude = *input.Longitude
	}

	marker.DiperbaruiPada = time.Now() // Perbarui timestamp DiperbaruiPada

	if err := tc.DB.Save(&marker).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update marker: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Marker updated successfully", "marker": marker})
}

// DeleteMarker adalah metode dari TravelController yang menghapus marker.
// Membutuhkan token JWT dan hanya bisa menghapus marker yang dimiliki pengguna.
func (tc *TravelController) DeleteMarker(c *gin.Context) {
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

	var marker models.Place
	// Jika role admin, izinkan hapus marker apapun; jika bukan, hanya marker miliknya
	role, _ := c.Get("role")
	var err error
	if role == "admin" {
		err = tc.DB.Where("id = ?", markerID).First(&marker).Error
	} else {
		err = tc.DB.Where("id = ? AND ditambahkan_oleh_user_id = ?", markerID, ownerUserID).First(&marker).Error
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
