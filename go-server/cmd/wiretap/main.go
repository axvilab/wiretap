package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	ingestPort := envOr("WIRETAP_INGEST_PORT", "9090")
	webPort := envOr("WIRETAP_WEB_PORT", "8080")

	// Ingest server
	ingestMux := http.NewServeMux()
	ingestMux.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})
	ingestMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy"}`)
	})

	// Web UI / API server
	webMux := http.NewServeMux()
	webMux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy"}`)
	})

	go func() {
		log.Printf("Ingest server listening on :%s", ingestPort)
		if err := http.ListenAndServe(":"+ingestPort, ingestMux); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("Web server listening on :%s", webPort)
	if err := http.ListenAndServe(":"+webPort, webMux); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
