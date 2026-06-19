package main

import (
	"bytes"
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
}

func TestProcess(t *testing.T) {
	payload := map[string]string{"key": "value"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/v1/process", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	handleProcess(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr.Code)
	}
}
