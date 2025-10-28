package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

// PLACE YOUR OPENWEATHER API KEY HERE
const OPENWEATHER_API_KEY = "ecd662cd0d71be48c2bd9832d183013b"

// WeatherResponse represents the OpenWeather API response structure
type WeatherResponse struct {
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Sys struct {
		Country string `json:"country"`
		Sunrise int64  `json:"sunrise"`
		Sunset  int64  `json:"sunset"`
	} `json:"sys"`
	Name string `json:"name"`
	Cod  int    `json:"cod"`
}

// ErrorResponse for API errors
type ErrorResponse struct {
	Error string `json:"error"`
}

// enableCORS middleware to handle CORS
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// weatherHandler handles weather API requests
func weatherHandler(w http.ResponseWriter, r *http.Request) {
	// Check if API key is set
	if OPENWEATHER_API_KEY == "YOUR_API_KEY_HERE" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "API key not configured. Please add your OpenWeather API key in main.go",
		})
		return
	}

	// Get city from query parameters
	city := r.URL.Query().Get("city")
	if city == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "City parameter is required",
		})
		return
	}

	// Build OpenWeather API URL
	apiURL := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric",
		url.QueryEscape(city),
		OPENWEATHER_API_KEY,
	)

	// Make request to OpenWeather API
	resp, err := http.Get(apiURL)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Failed to fetch weather data",
		})
		log.Printf("Error fetching weather data: %v", err)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Failed to read weather data",
		})
		log.Printf("Error reading response body: %v", err)
		return
	}

	// Handle non-200 status codes from OpenWeather API
	if resp.StatusCode != http.StatusOK {
		var apiError map[string]interface{}
		if err := json.Unmarshal(body, &apiError); err == nil {
			if message, ok := apiError["message"].(string); ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(resp.StatusCode)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error: fmt.Sprintf("Weather API error: %s", message),
				})
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "City not found or invalid request",
		})
		return
	}

	// Parse and validate response
	var weatherData WeatherResponse
	if err := json.Unmarshal(body, &weatherData); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Failed to parse weather data",
		})
		log.Printf("Error parsing JSON: %v", err)
		return
	}

	// Return weather data
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(weatherData)
	log.Printf("Successfully fetched weather for: %s", city)
}

// healthHandler for health checks
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"app":    "Tempify Weather API",
	})
}

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Register handlers
	http.HandleFunc("/weather", enableCORS(weatherHandler))
	http.HandleFunc("/health", enableCORS(healthHandler))

	// Start server
	log.Printf("🌡️  Tempify Weather API starting on port %s", port)
	log.Printf("📍 Weather endpoint: http://localhost:%s/weather?city=<city_name>", port)
	log.Printf("💚 Health check: http://localhost:%s/health", port)

	if OPENWEATHER_API_KEY == "YOUR_API_KEY_HERE" {
		log.Println("⚠️  WARNING: Please add your OpenWeather API key in main.go")
		log.Println("   Get your free API key at: https://openweathermap.org/api")
	}

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}