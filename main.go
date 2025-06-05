package main

import (
	"fmt"
	"log"
	"net/http" // Import net/http untuk StatusUnauthorized
	"os"
	"strings" // Import strings untuk AuthMiddleware

	"ulyngo/controllers" // Import controllers
	"ulyngo/models"      // Import models untuk AutoMigrate
	"ulyngo/utils"       // Import utils

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// AuthMiddleware adalah middleware JWT dasar untuk melindungi rute.
// Ini akan memverifikasi token JWT dari header Authorization.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort() // Menghentikan pemrosesan request lebih lanjut
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		claims, err := utils.VerifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set informasi pengguna dari claims ke konteks Gin
		// Ini dapat diakses oleh handler selanjutnya menggunakan c.Get("key")
		c.Set("userID", claims["sub"])
		c.Set("username", claims["username"])
		c.Set("role", claims["role"])

		c.Next() // Lanjutkan ke handler berikutnya
	}
}

func main() {
	// Muat variabel lingkungan dari file .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Inisialisasi koneksi database
	utils.ConnectDatabase()
	// Melakukan migrasi otomatis untuk model User dan Marker
	// Ini akan membuat tabel jika belum ada, atau memperbarui skema.
	// Jika ada perubahan tipe ID dari uint ke string (UUID), Anda mungkin perlu
	// menghapus tabel lama di database Anda untuk migrasi yang bersih saat pengembangan.
	utils.DB.AutoMigrate(&models.User{}, &models.Marker{})

	// Mengatur mode Gin (misal: debug, release)
	gin.SetMode(gin.ReleaseMode) // Disarankan untuk produksi
	router := gin.Default()

	// Middleware CORS untuk mengizinkan permintaan dari frontend
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // Sesuaikan untuk produksi
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH") // Tambahkan PATCH/PUT untuk update

		// Menangani preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent) // 204 No Content
			return
		}
		c.Next()
	})

	// Inisialisasi controller dengan dependensi database yang sudah terhubung
	authController := controllers.NewAuthController(utils.DB)
	travelController := controllers.NewTravelController(utils.DB)

	// Grup Rute Autentikasi
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/register", authController.Register)
		authRoutes.POST("/login", authController.Login)
	}

	// Rute Perjalanan (Beberapa rute bersifat publik, beberapa dilindungi)
	router.POST("/route", travelController.GetDirections) // Publik
	router.GET("/markers", travelController.GetMarkers)   // Publik (mendapatkan semua marker, tidak difilter berdasarkan user)

	// Rute CRUD Marker yang Dilindungi dengan AuthMiddleware
	protectedMarkerRoutes := router.Group("/api/markers")
	// Middleware untuk membatasi akses hanya untuk role admin
	adminOnly := func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Administrator privileges are required."})
			c.Abort()
			return
		}
		c.Next()
	}

	protectedMarkerRoutes.Use(AuthMiddleware(), adminOnly) // Hanya admin yang bisa akses
	{
		protectedMarkerRoutes.POST("/", travelController.AddMarker)         // Menambah marker
		protectedMarkerRoutes.PUT("/:id", travelController.UpdateMarker)    // Memperbarui marker berdasarkan ID
		protectedMarkerRoutes.DELETE("/:id", travelController.DeleteMarker) // Menghapus marker berdasarkan ID
	}

	// Mendapatkan port dari variabel lingkungan, default ke 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // Default port jika PORT tidak diset
	}

	fmt.Printf("Server running on port %s\n", port)
	router.Run(":" + port) // Menjalankan server Gin
}
