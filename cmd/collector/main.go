package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/karthik-minnikanti/cinnamon/internal/monitor"
)

var (
	serverURL = flag.String("server", "http://localhost:8080", "Server URL to send data to")
	interval  = flag.Duration("interval", time.Second, "Collection interval")
)

func main() {
	flag.Parse()

	// Initialize network monitor
	monitor := monitor.NewNetworkMonitor()

	// Create a channel to receive connection events
	connChan := make(chan *monitor.Connection, 100)
	monitor.SetConnectionChannel(connChan)

	// Start monitoring
	go monitor.Start()

	// Start sending data to server
	go sendDataToServer(connChan)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down collector...")
	monitor.Stop()
}

func sendDataToServer(connChan <-chan *monitor.Connection) {
	for conn := range connChan {
		data, err := json.Marshal(conn)
		if err != nil {
			log.Printf("Error marshaling connection data: %v", err)
			continue
		}

		resp, err := http.Post(*serverURL+"/api/connections", "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Printf("Error sending data to server: %v", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Server returned non-OK status: %d", resp.StatusCode)
		}
	}
}
