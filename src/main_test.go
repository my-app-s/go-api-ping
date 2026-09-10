// Copyright (C) 2025-2026 my-app-s
// Licensed under the GNU AGPLv3

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsRestrictedIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       net.IP
		expected bool
	}{
		{"Nil IP", nil, true},
		{"Loopback IPv4", net.IPv4(127, 0, 0, 1), true},
		{"Private Network 10.x", net.IPv4(10, 0, 0, 1), true},
		{"Private Network 192.168.x", net.IPv4(192, 168, 1, 1), true},
		{"Cloud Metadata IP", net.IPv4(169, 254, 169, 254), true},
		{"Unspecified IP", net.IPv4(0, 0, 0, 0), true},
		{"Public IP (Google DNS)", net.IPv4(8, 8, 8, 8), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRestrictedIP(tt.ip)
			if got != tt.expected {
				t.Errorf("isRestrictedIP(%v) = %v, expected %v", tt.ip, got, tt.expected)
			}
		})
	}
}

func TestCheckURL_EdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("Invalid Scheme", func(t *testing.T) {
		res := checkURL(ctx, "ftp://example.com")
		if res.Status != "DOWN" || res.Error != "forbidden or invalid URL scheme" {
			t.Errorf("Expected scheme error, got: %+v", res)
		}
	})

	t.Run("Malformed URL", func(t *testing.T) {
		res := checkURL(ctx, "://invalid-url")
		if res.Status != "DOWN" {
			t.Errorf("Expected DOWN, got: %+v", res)
		}
	})

	t.Run("Empty Hostname", func(t *testing.T) {
		res := checkURL(ctx, "http://")
		if res.Status != "DOWN" || res.Error != "empty hostname" {
			t.Errorf("Expected empty hostname error, got: %+v", res)
		}
	})

	t.Run("Explicit Localhost string", func(t *testing.T) {
		res := checkURL(ctx, "http://localhost:8080")
		if res.Status != "DOWN" || res.Error != "forbidden internal host" {
			t.Errorf("Expected forbidden internal host error, got: %+v", res)
		}
	})

	t.Run("IP literal that is restricted (127.0.0.1)", func(t *testing.T) {
		res := checkURL(ctx, "http://127.0.0.1")
		if res.Status != "DOWN" || res.Error != "forbidden internal IP" {
			t.Errorf("Expected forbidden internal IP error, got: %+v", res)
		}
	})

	t.Run("Unresolvable Domain", func(t *testing.T) {
		res := checkURL(ctx, "http://this-domain-definitely-does-not-exist-99999.local")
		if res.Status != "DOWN" {
			t.Errorf("Expected status DOWN, got: %+v", res)
		}
	})
}

func TestPingHandler_MethodsAndCORS(t *testing.T) {
	t.Run("CORS OPTIONS Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/ping", nil)
		rec := httptest.NewRecorder()

		pingHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
		if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("Expected CORS header, got %s", rec.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("Method Not Allowed (GET)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
		rec := httptest.NewRecorder()

		pingHandler(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", rec.Code)
		}
	})
}

func TestPingHandler_ValidationAndLimits(t *testing.T) {
	t.Run("Invalid JSON Body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/ping", bytes.NewBufferString("not-a-json"))
		rec := httptest.NewRecorder()

		pingHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", rec.Code)
		}
	})

	t.Run("Too Many URLs (> 50)", func(t *testing.T) {
		var urls []string
		for i := 0; i < 51; i++ {
			urls = append(urls, "http://example.com")
		}
		bodyBytes, _ := json.Marshal(PingRequest{URLs: urls})
		req := httptest.NewRequest(http.MethodPost, "/api/ping", bytes.NewBuffer(bodyBytes))
		rec := httptest.NewRecorder()

		pingHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", rec.Code)
		}
	})

	t.Run("Valid Batch Request Handling", func(t *testing.T) {
		payload := PingRequest{
			URLs: []string{
				"http://127.0.0.1",
				"http://this-domain-does-not-exist-99999.local",
			},
		}
		bodyBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/ping", bytes.NewBuffer(bodyBytes))
		rec := httptest.NewRecorder()

		pingHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}

		var results []PingResult
		if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})
}
