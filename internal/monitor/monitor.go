package monitor

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/karthik-minnikanti/cinnamon/internal/models"
	"github.com/karthik-minnikanti/cinnamon/internal/storage"
)

// Common service ports
var servicePorts = map[int]struct {
	ServiceType      models.ServiceType
	DatabaseType     models.DatabaseType
	MessageQueueType models.MessageQueueType
	Name             string
}{
	// Databases
	5432:  {models.ServiceTypeDatabase, models.DatabaseTypePostgreSQL, "", "PostgreSQL"},
	3306:  {models.ServiceTypeDatabase, models.DatabaseTypeMySQL, "", "MySQL"},
	27017: {models.ServiceTypeDatabase, models.DatabaseTypeMongoDB, "", "MongoDB"},
	6379:  {models.ServiceTypeDatabase, models.DatabaseTypeRedis, "", "Redis"},

	// Message Queues
	5672: {models.ServiceTypeMessageQueue, "", models.MessageQueueTypeRabbitMQ, "RabbitMQ"},
	9092: {models.ServiceTypeMessageQueue, "", models.MessageQueueTypeKafka, "Kafka"},

	// Other common services
	80:   {models.ServiceTypeAPI, "", "", "HTTP"},
	443:  {models.ServiceTypeAPI, "", "", "HTTPS"},
	8080: {models.ServiceTypeAPI, "", "", "HTTP-Alt"},
}

// NetworkMonitor tracks network connections and errors
type NetworkMonitor struct {
	storage  storage.Storage
	stop     chan struct{}
	wg       sync.WaitGroup
	connChan chan *models.Connection
}

// NewNetworkMonitor creates a new network monitor instance
func NewNetworkMonitor(storage storage.Storage) *NetworkMonitor {
	// Initialize random number generator with current time as seed
	rand.Seed(time.Now().UnixNano())

	return &NetworkMonitor{
		storage:  storage,
		stop:     make(chan struct{}),
		connChan: make(chan *models.Connection, 100),
	}
}

// Start begins monitoring network connections
func (m *NetworkMonitor) Start() {
	m.wg.Add(1)
	go m.monitorConnections()
}

// Stop gracefully stops the monitor
func (m *NetworkMonitor) Stop() {
	close(m.stop)
	m.wg.Wait()
}

func (m *NetworkMonitor) monitorConnections() {
	defer m.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stop:
			return
		case <-ticker.C:
			m.checkConnections()
		}
	}
}

func (m *NetworkMonitor) checkConnections() {
	cmd := exec.Command("netstat", "-an")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error running netstat: %v", err)
		return
	}

	// Keep track of connections seen in this check
	seenConnections := make(map[string]bool)

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "tcp") {
			continue
		}

		conn := m.parseConnection(line)
		if conn != nil {
			// Create a connection key without timestamp and random component
			connKey := fmt.Sprintf("%s:%d-%s:%d", conn.SourceIP, conn.SourcePort, conn.DestIP, conn.DestPort)

			// Skip if we've already seen this connection in this check
			if seenConnections[connKey] {
				continue
			}
			seenConnections[connKey] = true

			// Identify service type based on port
			if serviceInfo, ok := servicePorts[conn.DestPort]; ok {
				conn.ServiceType = serviceInfo.ServiceType
				conn.DatabaseType = serviceInfo.DatabaseType
				conn.MessageQueueType = serviceInfo.MessageQueueType
				conn.ServiceName = serviceInfo.Name
			}

			// Add additional metadata
			conn.Metadata = map[string]interface{}{
				"protocol":    "TCP",
				"detected_at": time.Now().Format(time.RFC3339),
			}

			// Only send to channel, don't store locally
			select {
			case m.connChan <- conn:
			default:
				log.Println("Connection channel full, dropping connection")
			}
		}
	}
}

func (m *NetworkMonitor) parseConnection(line string) *models.Connection {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return nil
	}

	// Skip header lines and non-TCP lines
	if !strings.Contains(line, "tcp") {
		return nil
	}

	// Parse local address (format: "192.168.1.36.52066" on macOS)
	localAddr := fields[3]
	localParts := strings.Split(localAddr, ".")
	if len(localParts) < 2 {
		return nil
	}

	// Handle IPv4 addresses with dots
	sourceIP := strings.Join(localParts[:len(localParts)-1], ".")
	sourcePort, err := strconv.Atoi(localParts[len(localParts)-1])
	if err != nil {
		return nil
	}

	// Parse remote address (format: "34.139.23.89.9092" on macOS)
	remoteAddr := fields[4]
	remoteParts := strings.Split(remoteAddr, ".")
	if len(remoteParts) < 2 {
		return nil
	}

	// Handle IPv4 addresses with dots
	destIP := strings.Join(remoteParts[:len(remoteParts)-1], ".")
	destPort, err := strconv.Atoi(remoteParts[len(remoteParts)-1])
	if err != nil {
		return nil
	}

	// Create connection object with timestamp-based ID and random component
	timestamp := time.Now()
	random := rand.Int63n(1000000) // Add a random number between 0 and 999999
	conn := &models.Connection{
		ID:               fmt.Sprintf("%s:%d-%s:%d-%d-%d", sourceIP, sourcePort, destIP, destPort, timestamp.UnixNano(), random),
		Timestamp:        timestamp,
		SourceIP:         sourceIP,
		SourcePort:       sourcePort,
		DestIP:           destIP,
		DestPort:         destPort,
		Protocol:         "TCP",
		ServiceType:      models.ServiceTypeOther,
		DatabaseType:     models.DatabaseTypeOther,
		MessageQueueType: models.MessageQueueTypeOther,
		Tags:             []string{"tcp", "network-monitor"},
	}

	return conn
}

func getProcessID(localAddr string) (int, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		// Use lsof on macOS which doesn't require root privileges
		cmd = exec.Command("lsof", "-i", "-n", "-P")
	case "linux":
		cmd = exec.Command("netstat", "-anp")
	case "windows":
		cmd = exec.Command("netstat", "-ano", "-p", "TCP")
	default:
		return 0, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to execute command: %v", err)
	}

	// Parse the output to get the process ID
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, localAddr) {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				// On macOS with lsof, PID is the second field
				// Format: "process PID user ..."
				pid := 0
				fmt.Sscanf(fields[1], "%d", &pid)
				if pid > 0 {
					return pid, nil
				}
			}
		}
	}

	// If we can't find the PID, return 0 instead of an error
	return 0, nil
}

func getProcessInfo(pid int) (string, string, error) {
	_, err := os.FindProcess(pid)
	if err != nil {
		return "", "", err
	}

	// Get process executable path
	exePath := ""
	switch runtime.GOOS {
	case "darwin", "linux":
		exePath, err = os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
		if err != nil {
			exePath = ""
		}
	case "windows":
		// On Windows, you might need to use different methods
		// This is a placeholder
		exePath = ""
	}

	processName := filepath.Base(exePath)
	if processName == "" {
		processName = fmt.Sprintf("process_%d", pid)
	}

	return processName, exePath, nil
}

func (m *NetworkMonitor) SetConnectionChannel(ch chan *models.Connection) {
	m.connChan = ch
}
