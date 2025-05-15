package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/karthik-minnikanti/cinnamon/internal/monitor"
	"github.com/karthik-minnikanti/cinnamon/internal/storage"
)

func main() {
	// Initialize storage
	db, err := storage.NewSQLiteStorage("network_monitor.db")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer db.Close()

	// Initialize network monitor
	monitor := monitor.NewNetworkMonitor(db)

	// Start monitoring
	go monitor.Start()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down network monitor...")
	monitor.Stop()
}
