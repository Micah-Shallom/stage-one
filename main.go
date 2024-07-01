package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"github.com/joho/godotenv"
	"strings"
)

type Geography struct {
	Location struct {
		Name   string `json:"name"`
		Region string `json:"region"`
	} `json:"location"`

	Current struct {
		Celsius   float64 `json:"temp_c"`
		Farenheit float64 `json:"temp_f"`
	} `json:"current"`
}

type Response struct {
	ClientIP string `json:"client_ip"`
	Location string `json:"location"`
	Greeting string `json:"greeting"`
}

func getClientIP(r *http.Request) string {
	// Check the X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// The X-Forwarded-For header can contain multiple IPs, use the first one
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	// Fall back to RemoteAddr
	ipPort := r.RemoteAddr
	ip := strings.Split(ipPort, ":")[0]
	return ip
}


func GetLocationInfo(apiKey, ipAddr string) (Geography, error) {
	var geoInfo Geography

	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, ipAddr)

	resp, err := http.Get(url)
	if err != nil {
		return geoInfo, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return geoInfo, err
	}

	if err := json.Unmarshal(body, &geoInfo); err != nil {
		return geoInfo, err
	}
	return geoInfo, nil
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	openWeatherAPIKey := os.Getenv("OPENWEATHER_API_KEY")
	visitorName := r.URL.Query().Get("visitor_name")
	clientIP := getClientIP(r)

	if clientIP == "127.0.0.1" || clientIP == "::1" {
		clientIP = "8.8.8.8"
	}

	temp, err := GetLocationInfo(openWeatherAPIKey, clientIP)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to fetch weather and location information: %v", err), http.StatusInternalServerError)
		return
	}

	greeting := fmt.Sprintf("Hello, %s! The temperature is %.1f degrees celsius in %s", visitorName, temp.Current.Celsius, temp.Location.Name)
	response := Response{
		ClientIP: clientIP,
		Location: temp.Location.Name,
		Greeting: greeting,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	http.HandleFunc("/", helloWorldHandler)
	http.HandleFunc("/api/hello", helloHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.ListenAndServe(":"+port, nil)
}
