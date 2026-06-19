package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "Go-Snapshot-Creator"})
}

var (
	snapshots = make(map[string]string)
	smu       sync.Mutex
)

func createSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	vol := r.URL.Query().Get("volume")
	if vol == "" {
		http.Error(w, "Missing volume", http.StatusBadRequest)
		return
	}
	id := "snap_" + vol
	smu.Lock()
	snapshots[id] = "completed"
	smu.Unlock()
	json.NewEncoder(w).Encode(map[string]string{"id": id, "status": "completed"})
}

func init() {
	http.HandleFunc("/snapshot", createSnapshotHandler)
}


func main() {
	http.HandleFunc("/health", healthHandler)
	fmt.Println("Starting Go-Snapshot-Creator on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
