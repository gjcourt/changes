// Package main is the entry point for changes, a jazz-standard chord
// progression database with transposition and Roman-numeral analysis.
package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"changes/internal/library"
	"changes/internal/server"
)

func main() {
	addr := env("ADDR", ":8080")
	webDir := env("WEB_DIR", "web")

	lib, err := library.Default()
	if err != nil {
		log.Fatalf("load library: %v", err)
	}
	log.Printf("loaded %d standards", len(lib.List()))

	srv := server.New(lib, webDir)

	log.Printf("listening on %s", addr)
	//nolint:gosec // simple server; no read/write timeout needed for LAN use
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
