package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/karthik-minnikanti/cinnamon/internal/api"
	"github.com/karthik-minnikanti/cinnamon/internal/storage"
)

var (
	port      = flag.String("port", "8080", "Server port")
	dbPath    = flag.String("db", "network_monitor.db", "Database path")
	staticDir = flag.String("static", "./static", "Static files directory")
)

func main() {
	flag.Parse()

	// Initialize storage
	db, err := storage.NewSQLiteStorage(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer db.Close()

	// Initialize API server
	server := api.NewServer(db, *staticDir)

	// Start HTTP server
	go func() {
		log.Printf("Starting server on port %s", *port)
		if err := http.ListenAndServe(":"+*port, server.Router()); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
}
