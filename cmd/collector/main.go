package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/karthik-minnikanti/cinnamon/internal/models"
	"github.com/karthik-minnikanti/cinnamon/internal/monitor"
	"github.com/karthik-minnikanti/cinnamon/internal/storage"
)

var (
	serverURL    = flag.String("server", "http://localhost:8080", "Server URL to send data to")
	interval     = flag.Duration("interval", time.Second, "Collection interval")
	serviceName  = flag.String("service", "", "Service name")
	host         = flag.String("host", "", "Host name")
	deploymentID = flag.String("deployment", "", "Deployment ID")
	environment  = flag.String("env", "production", "Environment")
	region       = flag.String("region", "", "Region")
)

func main() {
	flag.Parse()

	// Initialize storage
	storage, err := storage.NewSQLiteStorage("network_monitor.db")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize network monitor
	monitor := monitor.NewNetworkMonitor(storage)

	// Create a channel to receive connection events
	connChan := make(chan *models.Connection, 100)
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

func sendDataToServer(connChan <-chan *models.Connection) {
	// Keep track of recent connections to prevent duplicates
	recentConnections := make(map[string]time.Time)
	cleanupTicker := time.NewTicker(5 * time.Minute)
	defer cleanupTicker.Stop()

	for {
		select {
		case conn := <-connChan:
			// Create a connection key without timestamp and random component
			connKey := fmt.Sprintf("%s:%d-%s:%d", conn.SourceIP, conn.SourcePort, conn.DestIP, conn.DestPort)

			// Check if we've seen this connection recently (within last 5 minutes)
			if lastSeen, exists := recentConnections[connKey]; exists {
				if time.Since(lastSeen) < 5*time.Minute {
					continue // Skip if we've seen this connection recently
				}
			}

			// Update the last seen time
			recentConnections[connKey] = time.Now()

			// Add metadata
			conn.ServiceName = *serviceName
			conn.Host = *host
			conn.DeploymentID = *deploymentID
			conn.Environment = *environment
			conn.Region = *region

			// Add tags
			conn.Tags = append(conn.Tags,
				"network-monitor",
				*environment,
				*region,
				string(conn.ServiceType),
			)

			// Add service-specific tags
			if conn.ServiceType == models.ServiceTypeDatabase {
				conn.Tags = append(conn.Tags, string(conn.DatabaseType))
			} else if conn.ServiceType == models.ServiceTypeMessageQueue {
				conn.Tags = append(conn.Tags, string(conn.MessageQueueType))
			}

			// Add metadata
			conn.Metadata = map[string]interface{}{
				"collector_version": "1.0.0",
				"os":                "darwin",
				"start_time":        time.Now().Format(time.RFC3339),
				"service_type":      conn.ServiceType,
			}

			// Add service-specific metadata
			if conn.ServiceType == models.ServiceTypeDatabase {
				conn.Metadata["database_type"] = conn.DatabaseType
			} else if conn.ServiceType == models.ServiceTypeMessageQueue {
				conn.Metadata["queue_type"] = conn.MessageQueueType
			}

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

			// Check for successful response (200 OK or 201 Created)
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				log.Printf("Server returned unexpected status: %d", resp.StatusCode)
			}

		case <-cleanupTicker.C:
			// Clean up old connections from the map
			now := time.Now()
			for key, lastSeen := range recentConnections {
				if now.Sub(lastSeen) > 5*time.Minute {
					delete(recentConnections, key)
				}
			}
		}
	}
}
