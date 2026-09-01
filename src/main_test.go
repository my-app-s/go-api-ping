// Copyright (C) 2025-2026 my-app-s
// Licensed under the GNU AGPLv3

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsSafeURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://google.com", true},
		{"http://example.com/path", true},
		{"http://localhost:8080", false},
		{"http://127.0.0.1", false},
		{"http://10.0.0.1", false},
		{"http://192.168.1.1", false},
		{"file:///etc/passwd", false},
		{"ftp://example.com", false},
		{"not-a-url", false},
	}

	for _, tc := range tests {
		actual := isSafeURL(tc.url)
		if actual != tc.expected {
			t.Errorf("isSafeURL(%q) = %v; expected %v", tc.url, actual, tc.expected)
		}
	}
}

func TestPingHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	rec := httptest.NewRecorder()

	pingHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 Method Not Allowed, got %d", rec.Code)
	}
}

func TestPingHandler_PayloadTooLarge(t *testing.T) {
	// Создаем очень длинный массив URL, чтобы размер JSON превысил 64 КБ
	urls := make([]string, 2000)
	for i := range urls {
		urls[i] = "https://example.com/very/long/path/to/exceed/sixty/four/kilobytes/limit/in/request/body"
	}
	payload := PingRequest{URLs: urls}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/ping", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()

	pingHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request for large payload, got %d", rec.Code)
	}
}

func TestPingHandler_TooManyURLs(t *testing.T) {
	urls := make([]string, 51)
	for i := range urls {
		urls[i] = "https://example.com"
	}

	payload := PingRequest{URLs: urls}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/ping", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	pingHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request for >50 URLs, got %d", rec.Code)
	}
}

func TestPingHandler_ValidAndUnsafeURLs(t *testing.T) {
	payload := PingRequest{
		URLs: []string{
			"http://127.0.0.1", // Небезопасный URL (должен вернуть статус DOWN из-за SSRF фильтра)
			"https://invalid-domain-name-that-does-not-exist-12345.com", // Несуществующий хост
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/ping", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	pingHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", rec.Code)
	}

	var results []PingResult
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	for _, res := range results {
		if res.Status != "DOWN" {
			t.Errorf("Expected status DOWN for %s, got %s", res.URL, res.Status)
		}
		if res.Error == "" {
			t.Errorf("Expected an error message for %s, got empty string", res.URL)
		}
	}
}
