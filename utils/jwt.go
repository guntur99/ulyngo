package utils

import (
	"fmt" // Import fmt untuk error message
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5" // Menggunakan versi v5 dari library JWT
)

// jwtSecret adalah kunci rahasia yang digunakan untuk menandatangani dan memverifikasi JWT.
// Diambil dari variabel lingkungan JWT_SECRET.
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// GenerateToken membuat JWT baru untuk pengguna yang diberikan.
// Menerima userID (string), username (string), dan role (string) untuk dimasukkan ke dalam claims.
func GenerateToken(userID string, username string, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":      userID,                                // Subject: ID pengguna
		"username": username,                              // Username pengguna
		"role":     role,                                  // Role pengguna
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Kadaluarsa dalam 24 jam
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// VerifyToken memvalidasi string token JWT dan mengembalikan claims-nya.
func VerifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Memastikan algoritma penandatanganan adalah HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
