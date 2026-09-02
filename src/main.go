// Copyright (C) 2025-2026 my-app-s
// Licensed under the GNU AGPLv3

package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
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

// isRestrictedIP проверяет IP на локальные, приватные, служебные диапазоны и облачные метаданные
func isRestrictedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	// AWS / Cloud metadata IP: 169.254.169.254
	metadataIPv4 := net.IPv4(169, 254, 169, 254)
	if ip.Equal(metadataIPv4) {
		return true
	}

	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified()
}

func checkURL(targetURL string) PingResult {
	u, err := url.Parse(targetURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return PingResult{URL: targetURL, Status: "DOWN", Error: "forbidden or invalid URL scheme"}
	}

	hostname := u.Hostname()
	if hostname == "" {
		return PingResult{URL: targetURL, Status: "DOWN", Error: "empty hostname"}
	}

	// Безопасное извлечение порта с учетом IPv6 и дефолтных схем
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	if hostname == "localhost" {
		return PingResult{URL: targetURL, Status: "DOWN", Error: "forbidden internal host"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Явный резолв IP перед подключением (устранение DNS Rebinding / TOCTOU)
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		return PingResult{URL: targetURL, Status: "DOWN", Error: "failed to resolve host: " + err.Error()}
	}

	var targetIP net.IP
	for _, ipAddr := range ips {
		ip := ipAddr.IP
		if isRestrictedIP(ip) {
			return PingResult{URL: targetURL, Status: "DOWN", Error: "forbidden internal IP"}
		}
		targetIP = ip
		break // Берем первый валидный публичный IP
	}

	if targetIP == nil {
		return PingResult{URL: targetURL, Status: "DOWN", Error: "no valid public IP found"}
	}

	// Соединяемся напрямую с уже проверенным IP, исключая повторный DNS-запрос
	targetAddr := net.JoinHostPort(targetIP.String(), port)

	start := time.Now()
	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.Dial("tcp", targetAddr)
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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