package monitor

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/karthik-minnikanti/cinnamon/internal/models"
	"github.com/karthik-minnikanti/cinnamon/internal/storage"
)

// NetworkMonitor tracks network connections and errors
type NetworkMonitor struct {
	storage storage.Storage
	stop    chan struct{}
	wg      sync.WaitGroup
}

// NewNetworkMonitor creates a new network monitor instance
func NewNetworkMonitor(storage storage.Storage) *NetworkMonitor {
	return &NetworkMonitor{
		storage: storage,
		stop:    make(chan struct{}),
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
			if err := m.checkConnections(); err != nil {
				log.Printf("Error checking connections: %v", err)
			}
		}
	}
}

func (m *NetworkMonitor) checkConnections() error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd = exec.Command("netstat", "-an")
	case "windows":
		cmd = exec.Command("netstat", "-an")
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute netstat: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		conn, err := m.parseConnection(line)
		if err != nil {
			log.Printf("Error parsing connection: %v", err)
			continue
		}

		if conn != nil {
			if err := m.storage.StoreConnection(conn); err != nil {
				log.Printf("Error storing connection: %v", err)
			}
		}
	}

	return nil
}

func (m *NetworkMonitor) parseConnection(line string) (*models.Connection, error) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return nil, nil
	}

	// Extract local and remote addresses
	localAddr := fields[3]
	remoteAddr := fields[4]

	// Get process information
	pid, err := getProcessID(localAddr)
	if err != nil {
		return nil, err
	}

	processName, processPath, err := getProcessInfo(pid)
	if err != nil {
		return nil, err
	}

	// Determine connection state and potential errors
	state := "UNKNOWN"
	if len(fields) > 5 {
		state = fields[5]
	}

	// Check for common connection errors
	var connError models.ConnectionError
	switch state {
	case "ECONNREFUSED":
		connError = models.ErrConnRefused
	case "ECONNABORTED":
		connError = models.ErrConnAborted
	case "ECONNRESET":
		connError = models.ErrConnReset
	case "ETIMEDOUT":
		connError = models.ErrConnTimeout
	}

	return &models.Connection{
		Timestamp:   time.Now(),
		ProcessID:   pid,
		ProcessName: processName,
		ProcessPath: processPath,
		LocalAddr:   localAddr,
		RemoteAddr:  remoteAddr,
		Error:       connError,
		Protocol:    fields[0],
		State:       state,
	}, nil
}

func getProcessID(localAddr string) (int, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd = exec.Command("lsof", "-i", localAddr)
	case "windows":
		cmd = exec.Command("netstat", "-ano", "-p", "TCP")
	default:
		return 0, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Parse the output to get the process ID
	// This is a simplified version - you might need to adjust based on your OS
	lines := strings.Split(string(output), "\n")
	if len(lines) > 1 {
		fields := strings.Fields(lines[1])
		if len(fields) > 1 {
			pid := 0
			fmt.Sscanf(fields[1], "%d", &pid)
			return pid, nil
		}
	}

	return 0, fmt.Errorf("could not determine process ID")
}

func getProcessInfo(pid int) (string, string, error) {
	process, err := os.FindProcess(pid)
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
