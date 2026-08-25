package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var volumePattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)

type Snapshot struct {
	ID        string    `json:"id"`
	Volume    string    `json:"volume"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	mu        sync.RWMutex
	snapshots map[string]Snapshot
	maxItems  int
}

func newStore(maxItems int) *Store {
	if maxItems <= 0 {
		maxItems = 1000
	}
	return &Store{snapshots: make(map[string]Snapshot), maxItems: maxItems}
}

func (s *Store) Create(volume string) (Snapshot, error) {
	volume = strings.TrimSpace(volume)
	if !volumePattern.MatchString(volume) {
		return Snapshot{}, errors.New("invalid volume")
	}

	sum := sha256.Sum256([]byte(volume))
	id := "snap_" + hex.EncodeToString(sum[:8])

	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.snapshots[id]; ok {
		return existing, nil
	}
	if len(s.snapshots) >= s.maxItems {
		return Snapshot{}, errors.New("snapshot capacity reached")
	}

	snapshot := Snapshot{ID: id, Volume: volume, Status: "completed", CreatedAt: time.Now().UTC()}
	s.snapshots[id] = snapshot
	return snapshot, nil
}

func (s *Store) Get(id string) (Snapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snapshot, ok := s.snapshots[id]
	return snapshot, ok
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.snapshots)
}

type Server struct {
	store  *Store
	logger *slog.Logger
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.healthHandler)
	mux.HandleFunc("GET /ready", s.readyHandler)
	mux.HandleFunc("GET /metrics", s.metricsHandler)
	mux.HandleFunc("POST /api/v1/snapshots", s.createSnapshotHandler)
	mux.HandleFunc("GET /api/v1/snapshots/{id}", s.getSnapshotHandler)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "sky-snapshot"})
}

func (s *Server) readyHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) metricsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = fmt.Fprintf(w, "sky_snapshot_items %d\n", s.store.Count())
}

func (s *Server) createSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var body struct {
		Volume string `json:"volume"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	snapshot, err := s.store.Create(body.Volume)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "snapshot capacity reached" {
			status = http.StatusServiceUnavailable
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	s.logger.Info("snapshot created", "snapshot_id", snapshot.ID, "volume", snapshot.Volume)
	writeJSON(w, http.StatusCreated, snapshot)
}

func (s *Server) getSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !strings.HasPrefix(id, "snap_") || len(id) > 64 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid snapshot id"})
		return
	}

	snapshot, ok := s.store.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "snapshot not found"})
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func main() {
	maxItems := 1000
	if raw := os.Getenv("MAX_SNAPSHOTS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 100000 {
			panic("MAX_SNAPSHOTS must be an integer between 1 and 100000")
		}
		maxItems = parsed
	}

	addr := os.Getenv("BIND_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	server := &Server{store: newStore(maxItems), logger: logger}
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           server.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Info("starting snapshot service", "bind_addr", addr, "max_snapshots", maxItems)
	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
