package controllers

import (
	"bytes"         // Untuk Natural Language API
	"encoding/json" // Untuk JSON
	"io"
	"log"
	"net/http"
	"net/url" // Import package net/url
	"os"      // Untuk mengakses variabel lingkungan

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ulyngo/models" // Import model Anda

	"github.com/google/uuid" // Untuk UUID
)

type RouteController struct {
	DB *gorm.DB
}

func NewRouteController(db *gorm.DB) *RouteController {
	return &RouteController{DB: db}
}

// Struct untuk input rute dari frontend
type RouteRequest struct {
	Origin      string `json:"origin" binding:"required"`
	Destination string `json:"destination" binding:"required"`
	IsPublic    bool   `json:"is_public"` // Untuk menyimpan rute oleh user yang terautentikasi
}

// Struct untuk respons Directions API (diperluas untuk menyimpan ke DB)
type DirectionsAPIResponse struct {
	Routes []struct {
		Legs []struct {
			Distance struct {
				Text  string `json:"text"`
				Value int64  `json:"value"` // Jarak dalam meter
			} `json:"distance"`
			Duration struct {
				Text  string `json:"text"`
				Value int64  `json:"value"` // Durasi dalam detik
			} `json:"duration"`
			// Lokasi awal dan akhir dari leg pertama (penting untuk origin/destination lat/lng)
			StartLocation struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"start_location"`
			EndLocation struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"end_location"`
			Steps []struct {
				HtmlInstructions string `json:"html_instructions"`
				Distance         struct {
					Text string `json:"text"`
				} `json:"distance"`
				Duration struct {
					Text string `json:"text"`
				} `json:"duration"`
				Polyline struct {
					Points string `json:"points"`
				} `json:"polyline"`
			} `json:"steps"`
		} `json:"legs"`
		OverviewPolyline struct {
			Points string `json:"points"`
		} `json:"overview_polyline"`
	} `json:"routes"`
	Status string `json:"status"` // "OK", "ZERO_RESULTS", etc.
	// error_message akan muncul jika status bukan "OK"
	ErrorMessage string `json:"error_message,omitempty"`
	// GeocodedWaypoints juga bisa berguna untuk info lebih lanjut tentang origin/destination yang di-resolve
	GeocodedWaypoints []struct {
		GeocoderStatus string   `json:"geocoder_status"`
		PlaceID        string   `json:"place_id"`
		Types          []string `json:"types"`
	} `json:"geocoded_waypoints,omitempty"`
}

// GetDirections memanggil Google Directions API dari backend dan menyimpan rute ke database.
// Rute ini dilindungi otentikasi untuk mendapatkan 'userID' dari konteks.
func (tc *RouteController) GetDirections(c *gin.Context) {
	var req RouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	googleMapsAPIKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if googleMapsAPIKey == "" {
		log.Println("GOOGLE_MAPS_API_KEY not set")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API key not configured"})
		return
	}

	encodedOrigin := url.QueryEscape(req.Origin)
	encodedDestination := url.QueryEscape(req.Destination)

	// Bangun URL untuk Google Directions API
	// Mengganti nama variabel 'url' menjadi 'apiURL' untuk menghindari shadowing package net/url
	apiURL := "https://maps.googleapis.com/maps/api/directions/json?" +
		"origin=" + encodedOrigin +
		"&destination=" + encodedDestination +
		"&key=" + googleMapsAPIKey

	resp, err := http.Get(apiURL) // Menggunakan apiURL yang baru
	if err != nil {
		log.Printf("Error calling Google Directions API: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get directions from Google API"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading Directions API response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read directions response"})
		return
	}

	var directionsResp DirectionsAPIResponse
	if err := json.Unmarshal(body, &directionsResp); err != nil {
		log.Printf("Raw Directions API response: %s", string(body))
		log.Printf("Error unmarshaling Directions API response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse directions response"})
		return
	}

	if directionsResp.Status == "OK" && len(directionsResp.Routes) > 0 {
		// Ambil data dari leg pertama dari rute pertama
		firstLeg := directionsResp.Routes[0].Legs[0]

		// Dapatkan UserID dan Username dari konteks Gin
		userIDVal, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to save route. User ID not found in context."})
			return
		}
		parsedUserID, err := uuid.Parse(userIDVal.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format in context"})
			return
		}

		usernameVal, _ := c.Get("username") // Ambil username (opsional, untuk logging/debug)
		log.Printf("Route request from user: %s (ID: %s)", usernameVal, parsedUserID.String())

		// Siapkan data untuk model Route
		routeToSave := models.Route{
			Name:            req.Origin + " to " + req.Destination, // Nama default, bisa disesuaikan
			OriginText:      req.Origin,
			DestinationText: req.Destination,
			OriginLat:       firstLeg.StartLocation.Lat,
			OriginLng:       firstLeg.StartLocation.Lng,
			DestinationLat:  firstLeg.EndLocation.Lat,
			DestinationLng:  firstLeg.EndLocation.Lng,
			DistanceMeters:  firstLeg.Distance.Value,
			DurationSeconds: firstLeg.Duration.Value,
			UserID:          parsedUserID, // Menggunakan UserID dari konteks
			IsPublic:        req.IsPublic, // Gunakan is_public dari request
		}

		// Konversi seluruh respons Directions API ke json.RawMessage
		fullRouteDataJSON, err := json.Marshal(directionsResp)
		if err != nil {
			log.Printf("Error marshaling full directions response to JSON: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize route data"})
			return
		}
		routeToSave.RouteDataJSON = fullRouteDataJSON

		// Simpan rute ke database
		if err := tc.DB.Create(&routeToSave).Error; err != nil {
			log.Printf("Error saving route to database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save route"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Route calculated and saved successfully",
			"route":   routeToSave,    // Kembalikan objek rute yang disimpan
			"details": directionsResp, // Kembalikan detail directions API juga
		})
	} else {
		// Kirim status dan pesan error dari Google API jika tidak OK atau tidak ada rute
		statusMessage := directionsResp.Status
		if directionsResp.ErrorMessage != "" {
			statusMessage += ": " + directionsResp.ErrorMessage
		}
		c.JSON(http.StatusOK, gin.H{"status": directionsResp.Status, "message": "No routes found or API error: " + statusMessage})
	}
}

// Struct untuk input PlaceSearch dari frontend
type PlaceSearchRequest struct {
	Query        string `json:"query" binding:"required"`
	LocationBias string `json:"location_bias"` // Optional, e.g., "point:latitude,longitude"
}

// Struct untuk respons Places API (sederhana)
type PlacesAPIResponse struct {
	Results []struct { // Diubah dari 'Candidates' menjadi 'Results'
		PlaceID          string `json:"place_id"`
		Name             string `json:"name"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
		// Tambahkan properti lain yang mungkin ada di respons Text Search API jika diperlukan
		// Contoh: Types []string `json:"types"`
		//          OpeningHours struct { OpenNow bool `json:"open_now"` } `json:"opening_hours"`
		//          Rating      float64 `json:"rating"`
		//          PriceLevel  int `json:"price_level"`
		//          UserRatingsTotal int `json:"user_ratings_total"`
	} `json:"results"` // Diubah dari 'candidates' menjadi 'results'
	Status        string `json:"status"`
	ErrorMessage  string `json:"error_message,omitempty"`   // Tambahkan untuk error yang lebih jelas
	NextPageToken string `json:"next_page_token,omitempty"` // Untuk pagination
}

// SearchPlaces memanggil Google Places Text Search API dari backend
func (tc *RouteController) SearchPlaces(c *gin.Context) {
	var req PlaceSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	googleMapsAPIKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if googleMapsAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API key not configured"})
		return
	}

	encodedQuery := url.QueryEscape(req.Query)

	apiURL := "https://maps.googleapis.com/maps/api/place/textsearch/json?" +
		"query=" + encodedQuery +
		"&key=" + googleMapsAPIKey

	if req.LocationBias != "" {
		apiURL += "&locationbias=" + url.QueryEscape(req.LocationBias)
	}

	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("Error calling Google Places API: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search places from Google API"})
		return
	}
	defer resp.Body.Close()

	// Periksa status HTTP sebelum membaca body dan unmarshal
	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := io.ReadAll(resp.Body) // Baca body untuk logging
		if readErr != nil {
			log.Printf("Error reading Places API non-OK response body: %v", readErr)
		}
		log.Printf("Places API returned non-OK status: %d. Raw response: %s", resp.StatusCode, string(bodyBytes))
		c.JSON(resp.StatusCode, gin.H{"error": "Google Places API returned non-OK status", "details": string(bodyBytes)})
		return
	}

	body, err := io.ReadAll(resp.Body) // Menggunakan io.ReadAll
	if err != nil {
		log.Printf("Error reading Places API response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read places response"})
		return
	}

	var placesResp PlacesAPIResponse
	if err := json.Unmarshal(body, &placesResp); err != nil {
		log.Printf("Raw Places API response (unmarshal failed): %s", string(body))
		log.Printf("Error unmarshaling Places API response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse places response"})
		return
	}

	if placesResp.Status == "OK" {
		c.JSON(http.StatusOK, placesResp)
	} else {
		// Kirim status dan pesan error dari Google API jika tidak OK atau API error
		c.JSON(http.StatusOK, gin.H{"status": placesResp.Status, "message": "No places found or API error"})
	}
}

// Input untuk Natural Language API
type AnalyzeSentimentRequest struct {
	Document struct {
		Type    string `json:"type"` // "PLAIN_TEXT" atau "HTML"
		Content string `json:"content"`
	} `json:"document"`
	EncodingType string `json:"encodingType"` // "UTF8", "UTF16", "UTF32"
}

// Respons dari Natural Language API (disederhanakan)
type SentimentResponse struct {
	DocumentSentiment struct {
		Magnitude float64 `json:"magnitude"` // Kekuatan emosi
		Score     float64 `json:"score"`     // -1.0 (negatif) sampai 1.0 (positif)
	} `json:"documentSentiment"`
	// ... properti lain untuk entities, syntax, dll.
}

// AnalyzeSentiment memanggil Natural Language API untuk analisis sentimen
func (tc *RouteController) AnalyzeSentiment(c *gin.Context) {
	var req struct {
		ReviewText string `json:"review_text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nlAPIKey := os.Getenv("GOOGLE_NATURAL_LANGUAGE_API_KEY")
	if nlAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Natural Language API key not configured"})
		return
	}

	nlReq := AnalyzeSentimentRequest{
		Document: struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		}{
			Type:    "PLAIN_TEXT",
			Content: req.ReviewText,
		},
		EncodingType: "UTF8",
	}

	jsonReq, err := json.Marshal(nlReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal NL request"})
		return
	}

	client := &http.Client{}
	// Mengganti nama variabel 'url' menjadi 'apiURL' untuk menghindari shadowing package net/url
	apiURL := "https://language.googleapis.com/v1/documents:analyzeSentiment?key=" + nlAPIKey
	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonReq)) // Menggunakan apiURL yang baru
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create HTTP request for NL API"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Natural Language API"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read NL API response"})
		return
	}

	var sentimentResp SentimentResponse
	if err := json.Unmarshal(body, &sentimentResp); err != nil {
		log.Printf("Error unmarshaling NL API response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse NL API response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"review_text": req.ReviewText,
		"sentiment":   sentimentResp.DocumentSentiment,
	})
}
