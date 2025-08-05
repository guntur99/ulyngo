package controllers

import (
	"bytes"
	"context" // Import context
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2/google" // <-- TAMBAHKAN IMPORT INI
	"gorm.io/gorm"
)

// =================================================================================
// KODE YANG ADA (DIREFAKTOR) DAN STRUCTS BARU
// =================================================================================

type RouteController struct {
	DB *gorm.DB
}

func NewRouteController(db *gorm.DB) *RouteController {
	return &RouteController{DB: db}
}

// Structs untuk request dan response Google Maps API (masih sama)
type RouteRequest struct {
	Origin      string `json:"origin" binding:"required"`
	Destination string `json:"destination" binding:"required"`
	IsPublic    bool   `json:"is_public"`
}

type DirectionsAPIResponse struct {
	Routes []struct {
		Legs []struct {
			Distance struct {
				Text  string `json:"text"`
				Value int64  `json:"value"`
			} `json:"distance"`
			Duration struct {
				Text  string `json:"text"`
				Value int64  `json:"value"`
			} `json:"duration"`
			StartLocation struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"start_location"`
			EndLocation struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"end_location"`
			// ... (properti lain tidak diubah)
		} `json:"legs"`
		OverviewPolyline struct {
			Points string `json:"points"`
		} `json:"overview_polyline"`
	} `json:"routes"`
	Status            string `json:"status"`
	ErrorMessage      string `json:"error_message,omitempty"`
	GeocodedWaypoints []struct {
		PlaceID string `json:"place_id"`
	} `json:"geocoded_waypoints,omitempty"`
}

type PlaceSearchRequest struct {
	Query        string `json:"query" binding:"required"`
	LocationBias string `json:"location_bias"`
}

type PlacesAPIResponse struct {
	Results []struct {
		PlaceID          string `json:"place_id"`
		Name             string `json:"name"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
		Rating float32 `json:"rating,omitempty"`
	} `json:"results"`
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// == Structs BARU untuk Fitur Perencanaan Perjalanan ==

// Struct untuk input dari user
type TripPlanRequest struct {
	// Kalimat dari user, e.g., "pengen ke braga jajan cimol..."
	Query string `json:"query" binding:"required"`
	// Lokasi awal user, e.g., "Jakarta, Indonesia" atau "latitude,longitude"
	Origin string `json:"origin" binding:"required"`
}

// Struct untuk menampung hasil ekstraksi dari Vertex AI
type ExtractedTripInfo struct {
	Destination      string   `json:"destination"`
	StopsAlongTheWay []string `json:"stops_along_the_way"`
	ReturnTripPlan   string   `json:"return_trip_plan"`
}

// Struct untuk request ke Vertex AI Gemini
type VertexAIRequest struct {
	Contents struct {
		Role  string `json:"role"`
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

// Struct untuk response dari Vertex AI Gemini (disederhanakan)
type VertexAIResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Struct untuk respons akhir yang komprehensif
type FinalTripPlanResponse struct {
	Interpretation ExtractedTripInfo            `json:"interpretation"`
	MainRoute      DirectionsAPIResponse        `json:"main_route"`
	SuggestedStops map[string]PlacesAPIResponse `json:"suggested_stops"` // map[nama_stop]hasil_pencarian
	ReturnTripShop PlacesAPIResponse            `json:"return_trip_shop"`
}

// =================================================================================
// FUNGSI HELPER BARU (Refaktor dari Controller Lama)
// =================================================================================

// getDirectionsData adalah versi refaktor dari GetDirections
// Fungsi ini hanya mengambil data dari API dan mengembalikannya, tanpa menyimpan ke DB atau menulis respons HTTP.
func (tc *RouteController) getDirectionsData(origin, destination string) (*DirectionsAPIResponse, error) {
	googleMapsAPIKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if googleMapsAPIKey == "" {
		log.Println("ERROR: GOOGLE_MAPS_API_KEY environment variable not set.")
		return nil, fmt.Errorf("API key not configured")
	}

	apiURL := fmt.Sprintf("https://maps.googleapis.com/maps/api/directions/json?origin=%s&destination=%s&key=%s",
		url.QueryEscape(origin),
		url.QueryEscape(destination),
		googleMapsAPIKey)

	// === LANGKAH DEBUG 1: Cetak URL yang akan dipanggil ===
	// Ini membantu memastikan origin dan destination sudah benar.
	log.Printf("Calling Google Directions API with URL: %s", apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("ERROR: Failed to call Google Directions API: %v", err)
		return nil, fmt.Errorf("failed to call Google Directions API: %w", err)
	}
	defer resp.Body.Close()

	// === LANGKAH DEBUG 2: Cetak status HTTP dari respons ===
	log.Printf("Google Directions API returned HTTP status: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read directions response body: %v", err)
		return nil, fmt.Errorf("failed to read directions response: %w", err)
	}

	// === LANGKAH DEBUG 3: Selalu cetak body respons mentah ===
	// Ini adalah langkah paling penting. Kita bisa lihat pesan error dari Google.
	log.Printf("Raw Directions API response: %s", string(body))

	var directionsResp DirectionsAPIResponse
	if err := json.Unmarshal(body, &directionsResp); err != nil {
		log.Printf("ERROR: Failed to parse directions response JSON: %v", err)
		return nil, fmt.Errorf("failed to parse directions response: %w", err)
	}

	// === LANGKAH DEBUG 4: Cek status internal dari Google ===
	// Setelah berhasil di-parse, cek status logikanya.
	if directionsResp.Status != "OK" {
		errorMessage := fmt.Sprintf("Directions API returned status '%s'. Full error: %s", directionsResp.Status, directionsResp.ErrorMessage)
		log.Printf("ERROR: %s", errorMessage)
		return nil, fmt.Errorf(errorMessage)
	}

	log.Println("Successfully received and parsed directions.")
	return &directionsResp, nil
}

// searchPlacesData adalah versi refaktor dari SearchPlaces
// Hanya mencari dan mengembalikan data, tidak menulis respons HTTP.
func (tc *RouteController) searchPlacesData(query, locationBias string) (*PlacesAPIResponse, error) {
	googleMapsAPIKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if googleMapsAPIKey == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	apiURL := fmt.Sprintf("https://maps.googleapis.com/maps/api/place/textsearch/json?query=%s&key=%s",
		url.QueryEscape(query),
		googleMapsAPIKey)

	// Lokasi bias membantu API memberikan hasil yang lebih relevan dengan area tujuan
	if locationBias != "" {
		apiURL += "&location=" + url.QueryEscape(locationBias)
		apiURL += "&radius=100" // cari dalam radius 10km dari titik lokasi bias
	}

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call Google Places API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read places response: %w", err)
	}

	var placesResp PlacesAPIResponse
	if err := json.Unmarshal(body, &placesResp); err != nil {
		log.Printf("Raw Places API response: %s", string(body))
		return nil, fmt.Errorf("failed to parse places response: %w", err)
	}

	if placesResp.Status != "OK" && placesResp.Status != "ZERO_RESULTS" {
		return nil, fmt.Errorf("places API error: %s - %s", placesResp.Status, placesResp.ErrorMessage)
	}

	return &placesResp, nil
}

// extractTripDetailsFromVertexAI memanggil Vertex AI untuk NLU (VERSI DIPERBARUI DENGAN ROLE & MODEL YANG BENAR)
func (tc *RouteController) extractTripDetailsFromVertexAI(query string) (*ExtractedTripInfo, error) {
	// Variabel lingkungan untuk Vertex AI
	projectID := os.Getenv("GOOGLE_VERTEX_AI_PROJECT_ID")
	location := os.Getenv("GOOGLE_VERTEX_AI_LOCATION") // e.g., "us-central1"

	if projectID == "" || location == "" {
		return nil, fmt.Errorf("vertex AI environment variables not configured (GOOGLE_VERTEX_AI_PROJECT_ID, GOOGLE_VERTEX_AI_LOCATION)")
	}

	ctx := context.Background()
	scopes := []string{"https://www.googleapis.com/auth/cloud-platform"}

	tokenSource, err := google.DefaultTokenSource(ctx, scopes...)
	if err != nil {
		return nil, fmt.Errorf("failed to create token source from service account: %w", err)
	}

	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Prompt yang spesifik untuk meminta output JSON (tidak berubah)
	prompt := fmt.Sprintf(`
      Ekstrak dan inferensi informasi rencana perjalanan dari kalimat berikut ke dalam format JSON yang ketat. Fokus utama adalah menentukan moda transportasi yang paling mungkin dan preferensi rute.
      Kalimat: "%s"

      Aturan:
      1.  **destination**: Harus berupa nama lokasi tujuan yang spesifik. Sertakan kota atau negara jika memungkinkan.
      2.  **travel_mode**: Objek yang berisi detail tentang cara perjalanan.
          - **mode**: **Lakukan inferensi** untuk menentukan moda transportasi dengan logika prioritas berikut:
              a. **Eksplisit**: Jika pengguna menyebut "mobil", gunakan "driving". Jika menyebut "motor", "motoran", atau "touring", gunakan "motorcycle".
              b. **Implisit/Kontekstual (Sangat Penting)**: Jika pengguna menggunakan frasa yang sangat mengindikasikan sepeda motor di konteks Indonesia seperti "lewat jalan tikus", "cari rute alternatif cepat", "selap-selip", atau "hindari ganjil-genap" (karena motor kebal ganjil-genap), **simpulkan sebagai "motorcycle"** bahkan jika kata 'motor' tidak disebut.
              c. **Default**: Jika sama sekali tidak ada petunjuk eksplisit maupun implisit, gunakan "driving" sebagai nilai default yang paling aman.
          - **preferences**: Array string berisi preferensi rute. Ekstrak dari frasa seperti "jangan lewat tol" (menjadi "avoid_tolls"), "hindari jalan raya" (menjadi "avoid_highways").
      3.  **stops_along_the_way**: Array makanan, minuman, atau aktivitas singkat selama perjalanan.
      4.  **return_trip_plan**: String rencana untuk perjalanan pulang.
      5.  **Nilai Kosong**: Gunakan nilai kosong yang sesuai ( "", [], {} ) jika informasi tidak ditemukan.

      ---
      Contoh 1 (Eksplisit)
      Kalimat: "Rute motoran ke Puncak, tapi jangan lewat tol ya."
      Output JSON:
      {
        "destination": "Puncak, Bogor, Indonesia",
        "travel_mode": {
          "mode": "motorcycle",
          "preferences": ["avoid_tolls"]
        },
        "stops_along_the_way": [],
        "return_trip_plan": ""
      }
      ---
      Contoh 2 (Implisit/Inferensi)
      Kalimat: "Mau ke Kota Tua dari Bekasi, cariin jalan tikus dong biar cepet nyampe."
      Output JSON:
      {
        "destination": "Kota Tua, Jakarta, Indonesia",
        "travel_mode": {
          "mode": "motorcycle", // Disimpulkan dari frasa "jalan tikus"
          "preferences": []
        },
        "stops_along_the_way": [],
        "return_trip_plan": ""
      }
      ---
      Contoh 3 (Default)
      Kalimat: "Tolong dong rute ke Lembang, mau beli oleh-oleh bolu susu."
      Output JSON:
      {
        "destination": "Lembang, Bandung Barat, Indonesia",
        "travel_mode": {
          "mode": "driving", // Tidak ada petunjuk, maka default ke mobil
          "preferences": []
        },
        "stops_along_the_way": [],
        "return_trip_plan": "beli oleh-oleh bolu susu"
      }
      ---
	  Contoh 4
      Kalimat: "Aku mau ke Jalan Braga Bandung naik mobil, di jalan pengen jajan cimol sama thai tea. Pulangnya mau beli oleh-oleh bolu susu lembang."
      Output JSON:
      {
        "destination": "Jalan Braga, Bandung, Indonesia",
        "travel_mode": {
          "mode": "driving",
          "preferences": []
        },
        "stops_along_the_way": ["cimol", "thai tea"],
        "return_trip_plan": "beli oleh-oleh bolu susu lembang"
      }
      ---
      Contoh 5
      Kalimat: "Rute motoran ke Puncak, tapi jangan lewat tol ya. Pengen ngopi dulu di jalan."
      Output JSON:
      {
        "destination": "Puncak, Bogor, Indonesia",
        "travel_mode": {
          "mode": "motorcycle",
          "preferences": ["avoid_tolls"]
        },
        "stops_along_the_way": ["ngopi"],
        "return_trip_plan": ""
      }
      ---

      JSON output:
    `, query)

	// ======================================================================
	// === PERBAIKAN DI SINI: Tambahkan "role": "user" ===
	// ======================================================================
	reqPayload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user", // <-- FIELD WAJIB UNTUK API VERSI BARU
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	jsonReq, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Vertex AI request: %w", err)
	}

	// Menggunakan nama model yang valid dan stabil
	apiURL := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/gemini-2.0-flash-001:generateContent", location, projectID, location)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonReq))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request for Vertex AI: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Vertex AI API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Vertex AI response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Vertex AI API returned non-OK status: %d. Raw response: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("vertex AI API returned non-OK status: %d", resp.StatusCode)
	}

	// Parsing respons (tidak berubah)
	var vertexResp VertexAIResponse
	if err := json.Unmarshal(body, &vertexResp); err != nil || len(vertexResp.Candidates) == 0 {
		log.Printf("Raw Vertex AI response (unmarshal failed or no candidates): %s", string(body))
		return nil, fmt.Errorf("failed to parse Vertex AI response or no candidates found")
	}

	jsonText := vertexResp.Candidates[0].Content.Parts[0].Text
	jsonText = strings.TrimPrefix(jsonText, "```json")
	jsonText = strings.TrimSuffix(jsonText, "```")

	var extractedInfo ExtractedTripInfo
	if err := json.Unmarshal([]byte(jsonText), &extractedInfo); err != nil {
		log.Printf("Failed to unmarshal JSON from Vertex AI text: %s. Error: %v", jsonText, err)
		return nil, fmt.Errorf("failed to unmarshal JSON from Vertex AI text: %w", err)
	}

	return &extractedInfo, nil
}

// =================================================================================
// CONTROLLER UTAMA YANG BARU
// =================================================================================

// PlanTripFromQuery adalah endpoint utama untuk merencanakan perjalanan dari natural language.
func (tc *RouteController) PlanTripFromQuery(c *gin.Context) {
	var req TripPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// === LANGKAH 1: Ekstrak informasi dari kalimat menggunakan Vertex AI ===
	extractedInfo, err := tc.extractTripDetailsFromVertexAI(req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to understand query", "details": err.Error()})
		return
	}
	if extractedInfo.Destination == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not determine a destination from the query."})
		return
	}

	// === LANGKAH 2: Dapatkan rute utama ke tujuan ===
	mainRoute, err := tc.getDirectionsData(req.Origin, extractedInfo.Destination)
	// c.JSON(http.StatusInternalServerError, gin.H{"message 1": req.Origin, "message 2": extractedInfo.Destination, "message 3": mainRoute})
	// return
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get main route", "details": err.Error()})
		return
	}

	// Tentukan titik tengah rute sebagai lokasi bias untuk pencarian
	destLocation := mainRoute.Routes[0].Legs[0].EndLocation
	locationBiasString := fmt.Sprintf("%f,%f", destLocation.Lat, destLocation.Lng)

	// === LANGKAH 3: Cari setiap tempat singgah yang diinginkan ===
	suggestedStops := make(map[string]PlacesAPIResponse)
	for _, stopQuery := range extractedInfo.StopsAlongTheWay {
		placesResp, err := tc.searchPlacesData(stopQuery, locationBiasString)
		if err != nil {
			log.Printf("Could not search for stop '%s': %v", stopQuery, err)
			continue // Lanjutkan ke stop berikutnya jika ada error
		}
		suggestedStops[stopQuery] = *placesResp
	}

	// === LANGKAH 4: Cari toko untuk rencana perjalanan pulang ===
	var returnShopResp *PlacesAPIResponse
	if extractedInfo.ReturnTripPlan != "" {
		returnShopResp, err = tc.searchPlacesData(extractedInfo.ReturnTripPlan, locationBiasString)
		if err != nil {
			log.Printf("Could not search for return trip plan '%s': %v", extractedInfo.ReturnTripPlan, err)
		}
	}

	// === LANGKAH 5: Gabungkan semua hasil dan kirim sebagai respons ===
	finalResponse := FinalTripPlanResponse{
		Interpretation: *extractedInfo,
		MainRoute:      *mainRoute,
		SuggestedStops: suggestedStops,
	}
	if returnShopResp != nil {
		finalResponse.ReturnTripShop = *returnShopResp
	}

	c.JSON(http.StatusOK, finalResponse)
}
