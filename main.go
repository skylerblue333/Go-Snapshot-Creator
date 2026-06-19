// Go-Snapshot-Creator: High-performance Go Snapshot Creator service
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func handleProcess(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "processed",
		"received":  len(payload),
		"timestamp": time.Now().Unix(),
	})
}


func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"service":   "Go-Snapshot-Creator",
		"timestamp": time.Now().Unix(),
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/v1/process", handleProcess)
	log.Printf("Go-Snapshot-Creator running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
