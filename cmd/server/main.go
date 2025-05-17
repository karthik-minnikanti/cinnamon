package main

import (
	"flag"
	"log"

	"github.com/karthik-minnikanti/cinnamon/internal/api"
	"github.com/karthik-minnikanti/cinnamon/internal/storage"
)

var (
	port   = flag.String("port", "8080", "Server port")
	dbPath = flag.String("db", "network.db", "Database path")
)

func main() {
	flag.Parse()

	// Initialize storage
	store, err := storage.NewSQLiteStorage(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize server
	server := api.NewServer(store)

	// Start server
	addr := ":" + *port
	log.Printf("Starting server on %s", addr)
	if err := server.Start(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
