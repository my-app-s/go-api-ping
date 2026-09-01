// Copyright (C) 2025-2026 my-app-s
// Licensed under the GNU AGPLv3

package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type PingRequest struct {
	URLs []string `json:"urls"`
}

type PingResult struct {
	URL     string        `json:"url"`
	Status  string        `json:"status"`
	Latency time.Duration `json:"latency"`
	Error   string        `json:"error,omitempty"`
}

// Защита от SSRF: проверка схемы и блокировка локальных/приватных адресов
func isSafeURL(targetURL string) bool {
	u, err := url.Parse(targetURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}

	host := u.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return false
	}

	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() {
			return false
		}
	}

	return true
}

func checkURL(targetURL string) PingResult {
	if !isSafeURL(targetURL) {
		return PingResult{URL: targetURL, Status: "DOWN", Error: "forbidden or invalid URL"}
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		return PingResult{URL: targetURL, Status: "DOWN", Error: err.Error()}
	}

	host := u.Host
	if !strings.Contains(host, ":") {
		if u.Scheme == "https" {
			host += ":443"
		} else {
			host += ":80"
		}
	}

	// Дополнительная проверка IP после резолва домена
	ips, err := net.LookupIP(u.Hostname())
	if err != nil {
		return PingResult{URL: targetURL, Status: "DOWN", Error: "failed to resolve host"}
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() {
			return PingResult{URL: targetURL, Status: "DOWN", Error: "forbidden internal IP"}
		}
	}

	start := time.Now()
	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.Dial("tcp", host)
	latency := time.Since(start)

	if err != nil {
		return PingResult{
			URL:     targetURL,
			Status:  "DOWN",
			Latency: latency,
			Error:   err.Error(),
		}
	}
	defer conn.Close()

	return PingResult{
		URL:     targetURL,
		Status:  "UP",
		Latency: latency,
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ограничение размера входящего JSON до 64 КБ (защита от перегрузки памяти)
	r.Body = http.MaxBytesReader(w, r.Body, 1024*64)

	var req PingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body or payload too large", http.StatusBadRequest)
		return
	}

	if len(req.URLs) > 50 {
		http.Error(w, "Too many URLs to check at once (max 50)", http.StatusBadRequest)
		return
	}

	results := make([]PingResult, len(req.URLs))
	var wg sync.WaitGroup

	// Семафор для ограничения одновременных горутин (максимум 10 параллельных TCP-соединений)
	sem := make(chan struct{}, 10)

	for i, u := range req.URLs {
		wg.Add(1)
		go func(index int, targetURL string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[index] = checkURL(targetURL)
		}(i, u)
	}

	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("./static")))
	mux.HandleFunc("/api/ping", pingHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("========================================")
	log.Printf("🚀 Server is running and listening")
	log.Printf("Local container port: %s", port)
	log.Printf("URL inside container: http://localhost:%s", port)
	log.Printf("========================================")

	if err := http.ListenAndServe("0.0.0.0:"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
