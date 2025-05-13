package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"epever_exporter/src/metrics"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	baseURL   string
	config    []string
	namespace string
	verbose   bool
)

func logVerbose(format string, v ...interface{}) {
	if verbose {
		log.Printf(format, v...)
	}
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	ip := os.Getenv("DEVICE_IP")
	namespace = os.Getenv("PROM_NAMESPACE")
	if ip == "" || namespace == "" {
		log.Fatal("DEVICE_IP and PROM_NAMESPACE must be set in .env")
	}
	baseURL = fmt.Sprintf("http://%s/livejson", ip)
	verbose = strings.ToLower(os.Getenv("LOG_VERBOSE")) == "true"
}

func loadConfig() {
	data, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Failed to read config.json: %v", err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse config.json: %v", err)
	}
}

func extractNestedValue(data map[string]any, path string) (float64, bool) {
	parts := strings.Split(path, ".")
	var current any = data
	for _, part := range parts {
		if m, ok := current.(map[string]any); ok {
			current = m[part]
		} else {
			return 0, false
		}
	}
	if val, ok := current.(float64); ok {
		return val, true
	}
	return 0, false
}

func collect() {
	refreshSecondsStr := os.Getenv("REFRESH_SECONDS")
	if refreshSecondsStr == "" {
		refreshSecondsStr = "15"
	}
	refreshSeconds, err := strconv.Atoi(refreshSecondsStr)
	if err != nil || refreshSeconds < 1 {
		log.Printf("Invalid REFRESH_SECONDS value '%s', defaulting to 15", refreshSecondsStr)
		refreshSeconds = 15
	}

	for {
		resp, err := http.Get(baseURL)
		if err != nil {
			log.Printf("Error fetching data: %v", err)
			time.Sleep(time.Duration(refreshSeconds) * time.Second)
			continue
		}
		logVerbose("HTTP GET %s", baseURL)

		var root map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&root); err != nil {
			log.Printf("Failed to parse JSON: %v", err)
			resp.Body.Close()
			time.Sleep(time.Duration(refreshSeconds) * time.Second)
			continue
		}
		resp.Body.Close()

		// Top-level metrics (like Wifi_RSSI)
		for _, path := range config {
			if !strings.Contains(path, ".") {
				if val, ok := extractNestedValue(root, path); ok {
					metrics.UpdateMetric(path, val, "root")
				}
			}
		}

		// EP_* device metrics
		for k, v := range root {
			if !strings.HasPrefix(k, "EP_") {
				continue
			}
			deviceID := strings.TrimPrefix(k, "EP_")
			epData, ok := v.(map[string]any)
			if !ok {
				continue
			}
			for _, path := range config {
				if !strings.Contains(path, ".") {
					continue
				}
				if val, ok := extractNestedValue(epData, path); ok {
					metrics.UpdateMetric(path, val, deviceID)
				}
			}
		}

		time.Sleep(time.Duration(refreshSeconds) * time.Second)
	}
}

func main() {
	loadEnv()
	loadConfig()
	metrics.Init(config)

	go collect()

	port := os.Getenv("PORT")
	if port == "" {
		port = "9100" // fallback default
	}

	http.Handle("/metrics", promhttp.HandlerFor(metrics.GetCustomRegistry(), promhttp.HandlerOpts{}))
	log.Printf("Starting HTTP server on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
