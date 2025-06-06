package controllers

import (
	"net/http"
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
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"` // Ini adalah plaintext password dari request
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
		PasswordHash: string(hashedPassword), // Menyimpan hash password
		Role:         "admin",                // Menetapkan role default sebagai "user"
	}

	// Menyimpan user ke database menggunakan DB yang di-inject
	if err := ac.DB.Create(&user).Error; err != nil {
		// Menangani error jika username sudah ada atau error database lainnya
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user: " + err.Error()})
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

	// Membandingkan plaintext password dari input dengan PasswordHash yang disimpan
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Membuat JWT untuk pengguna yang berhasil login
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role) // Meneruskan ID (string) dan Role
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
