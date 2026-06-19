package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type ServiceState struct {
	mu        sync.RWMutex
	Processed int
	Domain    string
}

var state = &ServiceState{Domain: "creator"}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"domain":    state.Domain,
		"processed": state.Processed,
	})
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	state.mu.Lock()
	state.Processed++
	state.mu.Unlock()
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, `{"status":"processing"}`)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/process", handleProcess)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server starting on :8080")
	log.Fatal(server.ListenAndServe())
}
