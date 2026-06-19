package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]string
}

var store = Store{data: make(map[string]string)}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	var payload map[string]string
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	store.mu.Lock()
	for k, v := range payload {
		store.data[k] = v
	}
	store.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "version": "3.0.0"})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/process", handleProcess)
	mux.HandleFunc("/health", handleHealth)
	
	log.Println("Go-Snapshot-Creator running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
