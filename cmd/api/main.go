package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/danny-molnar/crossword/internal/api"
	"github.com/danny-molnar/crossword/internal/store"
	"github.com/danny-molnar/crossword/internal/tools"
)

func main() {
	wordlistPath := envOrDefault("CRYPTIC_WORDLIST_PATH", "wordlists/english.txt")
	wordlist, err := tools.LoadWordlist(wordlistPath)
	if err != nil {
		log.Fatalf("load word list: %v", err)
	}

	address := ":" + envOrDefault("PORT", "8080")
	server := &http.Server{
		Addr:              address,
		Handler:           api.NewRouter(store.NewMemoryStore(), wordlist),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("crossword API listening on %s", address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve API: %v", err)
	}
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
