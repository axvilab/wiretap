package config

import (
	"encoding/json"
	"os"
	"strconv"
)

type Config struct {
	StoragePath string `json:"storage_path"`
	WebPort     string `json:"web_port"`
	IngestPort  string `json:"ingest_port"`
	SMTPPort    string `json:"smtp_port"`
	RetentionDays int  `json:"retention_days"`
	BatchSize   int    `json:"batch_size"`
	FlushMs     int    `json:"flush_ms"`
}

func Load() *Config {
	c := &Config{
		StoragePath:   envOr("WIRETAP_STORAGE_PATH", "wiretap.duckdb"),
		WebPort:       envOr("WIRETAP_WEB_PORT", "8080"),
		IngestPort:    envOr("WIRETAP_INGEST_PORT", "9090"),
		SMTPPort:      envOr("WIRETAP_SMTP_PORT", "1025"),
		RetentionDays: envInt("WIRETAP_RETENTION_DAYS", 7),
		BatchSize:     envInt("WIRETAP_BATCH_SIZE", 100),
		FlushMs:       envInt("WIRETAP_FLUSH_MS", 500),
	}

	// Try loading from config file
	if path := envOr("WIRETAP_CONFIG", "wiretap.yml"); path != "" {
		if data, err := os.ReadFile(path); err == nil {
			json.Unmarshal(data, c)
		}
	}

	return c
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
