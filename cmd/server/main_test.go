package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func testServer() *Server {
	return &Server{store: newStore(4), logger: testLogger()}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHealthAndReadiness(t *testing.T) {
	s := testServer()
	for _, path := range []string{"/health", "/ready"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()
		s.routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s returned %d", path, rr.Code)
		}
	}
}

func TestCreateAndGetSnapshot(t *testing.T) {
	s := testServer()
	body := bytes.NewBufferString(`{"volume":"vol-123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/snapshots", body)
	rr := httptest.NewRecorder()
	s.routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create returned %d: %s", rr.Code, rr.Body.String())
	}

	var created Snapshot
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Volume != "vol-123" {
		t.Fatalf("unexpected snapshot: %+v", created)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots/"+created.ID, nil)
	getRR := httptest.NewRecorder()
	s.routes().ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("get returned %d", getRR.Code)
	}
}

func TestRejectsInvalidVolumeAndUnknownFields(t *testing.T) {
	s := testServer()
	cases := []string{
		`{"volume":"../../etc/passwd"}`,
		`{"volume":"ok","unexpected":true}`,
	}
	for _, raw := range cases {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/snapshots", bytes.NewBufferString(raw))
		s.routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for %s, got %d", raw, rr.Code)
		}
	}
}

func TestCapacityAndIdempotency(t *testing.T) {
	s := &Server{store: newStore(1), logger: testLogger()}
	first, err := s.store.Create("vol-a")
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.store.Create("vol-a")
	if err != nil || first.ID != second.ID {
		t.Fatalf("expected idempotent create")
	}
	if _, err := s.store.Create("vol-b"); err == nil {
		t.Fatal("expected capacity error")
	}
}

func TestStoreConcurrentCreate(t *testing.T) {
	store := newStore(100)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = store.Create("shared-volume")
		}()
	}
	wg.Wait()
	if store.Count() != 1 {
		t.Fatalf("expected one idempotent snapshot, got %d", store.Count())
	}
}
