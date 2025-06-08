package controllers

import (
	"encoding/json"
	"net/http"
	"time"
	"ulyngo/models"
	"ulyngo/utils" // Pastikan utils diimpor untuk GenerateToken

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm" // Import gorm untuk akses database
)

// AuthController struct akan menampung dependensi database
type AuthController struct {
	DB *gorm.DB
}

// NewAuthController adalah konstruktor untuk AuthController.
// Menerima instance GORM DB untuk dependency injection.
func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{DB: db}
}

// RegisterInput adalah struktur untuk data yang diterima saat registrasi.
type RegisterInput struct {
	Username     string     `json:"username" binding:"required"`
	Password     string     `json:"password" binding:"required"` // Ini adalah plaintext password dari request
	Email        string     `json:"email" binding:"required"`    // Email pengguna, tidak null
	Whatsapp     *string    `json:"whatsapp"`                    // Nomor WhatsApp, bisa null
	LastActiveAt *time.Time `json:"last_active_at"`              // Waktu terakhir aktif, bisa null
}

type UserActivityLog struct {
	UserID       string          `gorm:"type:uuid;not null" json:"user_id"`
	ActivityType string          `gorm:"not null" json:"activity_type"`  // Tipe aktivitas (misal: 'register', 'login')
	TargetID     *string         `gorm:"type:uuid" json:"target_id"`     // ID entitas target (marker/route), bisa null
	ActivityData json.RawMessage `gorm:"type:text" json:"activity_data"` // Data aktivitas tambahan (JSONB)
	Timestamp    time.Time       `gorm:"autoCreateTime" json:"timestamp"`
}

// Register adalah metode dari AuthController yang menangani pendaftaran pengguna baru.
func (ac *AuthController) Register(c *gin.Context) {
	var input RegisterInput
	// Mem-bind JSON request body ke struct RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hashing password menggunakan bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Membuat instance User baru tanpa Email, dan role default sebagai "user"
	user := models.User{
		Username:     input.Username,
		Password:     string(hashedPassword), // Menyimpan hash password
		Email:        input.Email,            // Menyimpan email yang diterima dari input
		Role:         "admin",                // Menetapkan role default sebagai "user"
		Whatsapp:     input.Whatsapp,         // Menyimpan nomor WhatsApp jika ada
		LastActiveAt: input.LastActiveAt,     // Menyimpan waktu terakhir aktif jika ada
	}

	// Menyimpan user ke database menggunakan DB yang di-inject
	if err := ac.DB.Create(&user).Error; err != nil {
		// Menangani error jika username sudah ada atau error database lainnya
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user: " + err.Error()})
		return
	}

	activityLog := UserActivityLog{
		UserID:       user.ID.String(),
		ActivityType: "register",
		TargetID:     nil,                           // Tidak ada target ID untuk registrasi
		ActivityData: json.RawMessage([]byte("{}")), // Tidak ada data aktivitas tambahan untuk registrasi
		Timestamp:    time.Now(),
	}
	if err := ac.DB.Create(&activityLog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log activity: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// LoginInput adalah struktur untuk data yang diterima saat login.
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login adalah metode dari AuthController yang menangani proses login pengguna.
func (ac *AuthController) Login(c *gin.Context) {
	var input LoginInput
	// Mem-bind JSON request body ke struct LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	// Mencari pengguna berdasarkan username di database
	if err := ac.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Membandingkan plaintext password dari input dengan Password yang disimpan
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Membuat JWT untuk pengguna yang berhasil login
	token, err := utils.GenerateToken(user.ID.String(), user.Username, user.Role) // Meneruskan ID (string) dan Role
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Log aktivitas login
	activityLog := UserActivityLog{
		UserID:       user.ID.String(), // Menggunakan ID pengguna sebagai string
		ActivityType: "login",
		TargetID:     nil,
		ActivityData: json.RawMessage([]byte("{}")),
		Timestamp:    time.Now(),
	}
	if err := ac.DB.Create(&activityLog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log activity: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": gin.H{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	}})
	c.Set("userID", user.ID)         // Set userID ke konteks Gin untuk akses di middleware
	c.Set("username", user.Username) // Set username ke konteks Gin
	c.Set("role", user.Role)         // Set role ke konteks Gin
	c.Next()                         // Lanjutkan ke handler berikutnya
}
