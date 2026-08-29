package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/danny-molnar/cryptic-foundry/internal/api"
	"github.com/danny-molnar/cryptic-foundry/internal/store"
	"github.com/danny-molnar/cryptic-foundry/internal/tools"
)

func main() {
	wordlistPath := envOrDefault("CRYPTIC_WORDLIST_PATH", "wordlists/english.txt")
	wordlist, err := tools.LoadWordlist(wordlistPath)
	if err != nil {
		log.Fatalf("load word list: %v", err)
	}

	documentStore, err := store.OpenSQLiteDocumentStore(databasePath())
	if err != nil {
		log.Fatalf("open document store: %v", err)
	}
	defer documentStore.Close()
	memoryStore := store.NewMemoryStore()
	memoryStore.Documents = documentStore

	address := ":" + envOrDefault("PORT", "8080")
	server := &http.Server{
		Addr:              address,
		Handler:           api.NewRouter(memoryStore, wordlist),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Cryptic Foundry API listening on %s", address)
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

func databasePath() string {
	if value := os.Getenv("CRYPTIC_FOUNDRY_DB_PATH"); value != "" {
		return value
	}
	if value := os.Getenv("CROSSWORD_DB_PATH"); value != "" {
		return value
	}
	// Continue using an existing pre-rename database rather than hiding drafts.
	if _, err := os.Stat("crossword.db"); err == nil {
		return "crossword.db"
	}
	return "cryptic-foundry.db"
}
