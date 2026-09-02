// Copyright (C) 2025-2026 my-app-s
// Licensed under the GNU AGPLv3

package main

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsRestrictedIP(t *testing.T) {
	tests := []struct {
		ip       net.IP
		expected bool
	}{
		{net.ParseIP("8.8.8.8"), false},
		{net.ParseIP("1.1.1.1"), false},
		{net.ParseIP("127.0.0.1"), true},
		{net.ParseIP("10.0.0.1"), true},
		{net.ParseIP("192.168.1.1"), true},
		{net.ParseIP("169.254.169.254"), true},
		{nil, true},
	}

	for _, tc := range tests {
		actual := isRestrictedIP(tc.ip)
		if actual != tc.expected {
			t.Errorf("isRestrictedIP(%v) = %v; expected %v", tc.ip, actual, tc.expected)
		}
	}
}

func TestCheckURL_InvalidScheme(t *testing.T) {
	result := checkURL("file:///etc/passwd")
	if result.Status != "DOWN" || result.Error == "" {
		t.Errorf("Expected DOWN status and error for invalid scheme, got status=%s error=%s", result.Status, result.Error)
	}
}

func TestCheckURL_Localhost(t *testing.T) {
	result := checkURL("http://localhost:8080")
	if result.Status != "DOWN" || result.Error == "" {
		t.Errorf("Expected DOWN status and error for localhost, got status=%s error=%s", result.Status, result.Error)
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
			"http://127.0.0.1",
			"https://invalid-domain-name-that-does-not-exist-12345.com",
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