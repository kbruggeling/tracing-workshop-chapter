package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"api-service-2/internal/data"
)

func main() {
	// Set up HTTP handlers
	http.HandleFunc("/api/data", dataHandler)
	http.HandleFunc("/health", healthHandler)

	fmt.Println("API Service 2 is running on port 8080")
	http.ListenAndServe(":8080", nil)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "api-service-2"})
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	// Use standard HTTP client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("GET", "http://api-service-3:8080/api/data", nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Call API 3
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to call API 3", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read response from API 3 using internal package
	body, err := data.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response from API 3", http.StatusInternalServerError)
		return
	}

	// Return the data from API 3
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
