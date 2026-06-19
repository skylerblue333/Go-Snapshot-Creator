package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	handleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", resp["status"])
	}
}

func TestProcess(t *testing.T) {
	initial := state.Processed
	req, _ := http.NewRequest("POST", "/process", nil)
	rr := httptest.NewRecorder()
	handleProcess(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("Expected 202 Accepted, got %d", rr.Code)
	}

	if state.Processed != initial+1 {
		t.Errorf("Expected state to increment")
	}
}
