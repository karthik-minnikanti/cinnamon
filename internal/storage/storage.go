package storage

import (
	"github.com/karthik-minnikanti/cinnamon/internal/models"
)

// Storage defines the interface for storing connection data
type Storage interface {
	// StoreConnection saves a connection event
	StoreConnection(conn *models.Connection) error

	// GetConnections retrieves connection events with optional filters
	GetConnections(limit int, offset int) ([]*models.Connection, error)

	// GetConnectionsByError retrieves connections with specific error type
	GetConnectionsByError(error models.ConnectionError) ([]*models.Connection, error)

	// GetConnectionsByProcess retrieves connections for a specific process
	GetConnectionsByProcess(processID int) ([]*models.Connection, error)

	// Close closes the storage connection
	Close() error
}
