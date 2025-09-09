package main

import (
	"fmt"
	"log"
	"net/http" // Import net/http untuk StatusUnauthorized
	"os"
	"strings" // Import strings untuk AuthMiddleware

	"ulyngo/controllers" // Import controllers
	"ulyngo/db/seeders"  // Import seeders untuk seeding data awal
	"ulyngo/models"      // Import models untuk AutoMigrate
	"ulyngo/utils"       // Import utils

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware adalah middleware JWT dasar untuk melindungi rute.
// Ini akan memverifikasi token JWT dari header Authorization.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// --- PERBAIKAN CORS PREFLIGHT DI SINI ---
		// Jika metode permintaan adalah OPTIONS, langsung lewati ke handler berikutnya (atau CORS middleware Gin).
		// Permintaan OPTIONS tidak boleh diautentikasi.
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		// --- AKHIR PERBAIKAN ---

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
		c.Set("email", claims["email"])
		c.Set("role", claims["role"])

		c.Next() // Lanjutkan ke handler berikutnya
	}
}

// DBRefresh menghapus semua tabel yang terkait dengan model dan kemudian melakukan AutoMigrate.
// Ini SANGAT berisiko untuk produksi karena akan MENGHILANGKAN SEMUA DATA.
// Gunakan HANYA untuk lingkungan pengembangan/pengujian.
func DBRefresh(db *gorm.DB) {
	log.Println("Starting database refresh (dropping all tables)...")
	// Urutan penghapusan tabel penting jika ada foreign key constraints
	// Tabel yang memiliki foreign key ke tabel lain harus dihapus terlebih dahulu
	err := db.Migrator().DropTable(
		&models.MarkerCategory{},
		&models.MarkerTag{},
		&models.User{},
	)
	if err != nil {
		log.Fatalf("Failed to drop tables: %v", err)
	}
	log.Println("All tables dropped successfully.")

	log.Println("Running AutoMigrate after refresh...")
	db.AutoMigrate(
		&models.User{},
		&models.MarkerCategory{},
		&models.MarkerTag{},
	)
	log.Println("AutoMigrate completed after refresh.")
}

func main() {
	// START LOCAL MODE
	// Muat variabel lingkungan dari file .env
	// if err := godotenv.Load(); err != nil {
	// 	log.Printf("Warning: Error loading .env file: %v", err)
	// }

	// os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

	// //Create Google Cloud client
	// ctx := context.Background()
	// client, err := storage.NewClient(ctx)
	// if err != nil {
	// 	log.Fatalf("Failed to create client: %v", err)
	// }
	// defer client.Close()

	// END LOCAL MODE

	// credsJSON := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON")
	// if credsJSON == "" {
	// 	log.Fatal("❌ GOOGLE_APPLICATION_CREDENTIALS_JSON not found")
	// }

	// // Inisialisasi client dengan credentials dari JSON string
	// ctx := context.Background()
	// client, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(credsJSON)))
	// if err != nil {
	// 	log.Fatalf("❌ Failed to create client: %v", err)
	// }
	// defer client.Close()

	// log.Println("✅ GCS client created successfully")

	// Inisialisasi koneksi database
	utils.ConnectDatabase()
	// Melakukan migrasi otomatis untuk model User dan Marker
	// Ini akan membuat tabel jika belum ada, atau memperbarui skema.
	// Jika ada perubahan tipe ID dari uint ke string (UUID), Anda mungkin perlu
	// menghapus tabel lama di database Anda untuk migrasi yang bersih saat pengembangan.
	// Cek apakah ada argumen --seed
	if len(os.Args) > 1 && os.Args[1] == "--seed" {
		// Jalankan refresh database jika --seed diberikan
		DBRefresh(utils.DB) // Hapus dan recreate tabel
		seeders.RunAllSeeders(utils.DB)
		log.Println("Seeding complete. Server will now start with seeded data.")
		// Jika Anda hanya ingin seeding tanpa menjalankan server, uncomment baris di bawah:
		// os.Exit(0)
	} else {
		// Jika tidak ada argumen --seed, hanya lakukan AutoMigrate
		log.Println("Running AutoMigrate...")
		utils.DB.AutoMigrate(
			&models.User{},
			&models.Preference{},
			&models.MarkerCategory{},
			&models.MarkerTag{},
			&models.MarkerHasTag{},
			&models.Marker{},
			&models.MarkerImage{},
			&models.MarkerReview{},
			&models.Route{},
			&models.UserActivityLog{},
		)
		log.Println("AutoMigrate completed.")
	}

	// Mengatur mode Gin (misal: debug, release)
	gin.SetMode(gin.ReleaseMode) // Disarankan untuk produksi
	router := gin.Default()

	// Dapatkan FRONTEND_ORIGIN dari variabel lingkungan.
	// Ini adalah pendekatan yang direkomendasikan untuk produksi agar menentukan asal yang diizinkan.
	// Untuk pengembangan, Anda bisa mengaturnya ke "http://localhost:3000" atau sejenisnya.
	// Jika tidak disetel, akan default ke "*" tetapi perhatikan implikasinya dengan AllowCredentials.
	frontendOrigin := os.Getenv("FRONTEND_ORIGIN")
	if frontendOrigin == "" {
		frontendOrigin = "*" // Default ke semua asal untuk kenyamanan dalam pengembangan, tetapi berhati-hatilah.
		log.Println("WARNING: FRONTEND_ORIGIN environment variable not set. Using '*' for CORS. THIS IS NOT RECOMMENDED FOR PRODUCTION.")
	} else if frontendOrigin == "*" && os.Getenv("GIN_MODE") != "debug" { // Peringatan lebih keras jika di mode non-debug
		log.Println("CRITICAL WARNING: FRONTEND_ORIGIN is set to '*' with Access-Control-Allow-Credentials 'true'. This is a security risk and may cause CORS errors in production browsers. Please specify a concrete origin for production.")
	}

	// Middleware CORS untuk mengizinkan permintaan dari frontend
	router.Use(func(c *gin.Context) {
		// Set asal spesifik yang diizinkan.
		c.Writer.Header().Set("Access-Control-Allow-Origin", frontendOrigin)
		// Access-Control-Allow-Credentials harus true jika frontend mengirim cookie atau Authorization header
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH") // Tambahkan PATCH/PUT untuk update

		// Menangani preflight OPTIONS request
		// Browser mengirim OPTIONS request terlebih dahulu untuk memeriksa apakah permintaan diizinkan
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent) // Mengirim 204 No Content untuk preflight berhasil
			return
		}
		c.Next()
	})

	// Inisialisasi controller dengan dependensi database yang sudah terhubung
	authController := controllers.NewAuthController(utils.DB)
	markerController := controllers.NewMarkerController(utils.DB)
	markerCategoryController := controllers.NewMarkerCategoryController(utils.DB)
	markerTagController := controllers.NewMarkerTagController(utils.DB)
	routeController := controllers.NewRouteController(utils.DB)

	// Grup Rute Autentikasi
	authRoutes := router.Group("/api/auth")
	{
		authRoutes.POST("/register", authController.Register)
		authRoutes.POST("/login", authController.Login)
	}

	// Rute Perjalanan (Beberapa rute bersifat publik, beberapa dilindungi)
	// router.POST("/api/routes", routeController.GetDirections)                       // Publik
	router.GET("/api/markers", markerController.GetMarkers)                         // Publik (mendapatkan semua marker, tidak difilter berdasarkan user)
	router.GET("/api/marker/categories", markerCategoryController.GetAllCategories) // Publik (mendapatkan semua kategori marker)
	router.GET("/api/marker/tags", markerTagController.GetAllTags)                  // Publik (mendapatkan semua tag marker)

	// Rute CRUD Marker yang Dilindungi dengan AuthMiddleware
	protectedMarkerRoutes := router.Group("/api/markers")
	protectedMarkerCategoriesRoutes := router.Group("/api/marker/categories")
	protectedMarkerTagsRoutes := router.Group("/api/marker/tags")
	protectedServicesRoutes := router.Group("/api")
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

	// Grup Rute yang Dilindungi (membutuhkan otentikasi)
	protectedServicesRoutes.Use(AuthMiddleware())
	{
		// Rute rute (termasuk penyimpanan ke DB, sekarang dilindungi)
		// protectedServicesRoutes.POST("/routes", routeController.GetDirections)               // Pindahkan ke protectedRoutes
		// protectedServicesRoutes.POST("/places/search", routeController.SearchPlaces)         // Pindahkan ke protectedRoutes
		// protectedServicesRoutes.POST("/analyze-sentiment", routeController.AnalyzeSentiment) // Pindahkan ke protectedRoutes
		protectedServicesRoutes.POST("/plan-trip", routeController.PlanTripFromQuery)
	}

	protectedMarkerRoutes.Use(AuthMiddleware(), adminOnly)
	{
		// Mengubah rute POST dari "/" menjadi "" untuk menghilangkan trailing slash
		protectedMarkerRoutes.POST("", markerController.AddMarker)          // Menambah marker
		protectedMarkerRoutes.PUT("/:id", markerController.UpdateMarker)    // Memperbarui marker berdasarkan ID
		protectedMarkerRoutes.DELETE("/:id", markerController.DeleteMarker) // Menghapus marker berdasarkan ID
	}

	// Rute Marker Categories
	protectedMarkerCategoriesRoutes.Use(AuthMiddleware(), adminOnly)
	{
		protectedMarkerCategoriesRoutes.POST("", markerCategoryController.CreateCategory)
		protectedMarkerCategoriesRoutes.PUT("/:id", markerCategoryController.UpdateCategory)
		protectedMarkerCategoriesRoutes.DELETE("/:id", markerCategoryController.DeleteCategory)
	}

	// Rute Marker Tags
	protectedMarkerTagsRoutes.Use(AuthMiddleware(), adminOnly)
	{
		protectedMarkerTagsRoutes.POST("", markerTagController.CreateTag)
		protectedMarkerTagsRoutes.PUT("/:id", markerTagController.UpdateTag)
		protectedMarkerTagsRoutes.DELETE("/:id", markerTagController.DeleteTag)
	}

	// Mendapatkan port dari variabel lingkungan, default ke 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // Default port jika PORT tidak diset
	}

	fmt.Printf("Server running on port %s\n", port)
	router.Run(":" + port) // Menjalankan server Gin
}
