package storage

import (
	"time"

	"github.com/karthik-minnikanti/cinnamon/internal/models"
)

// Storage defines the interface for storing and retrieving connection data
type Storage interface {
	// Basic CRUD operations
	StoreConnection(conn *models.Connection) error
	GetConnections(service, errorType, environment, search string) ([]*models.Connection, error)
	GetConnectionByID(id string) (*models.Connection, error)

	// Statistics and analytics
	GetStats(startTime, endTime time.Time) (*models.ConnectionStats, error)
	GetServices() ([]string, error)
	GetErrors() ([]string, error)
	GetEnvironments() ([]string, error)

	// Cleanup
	Close() error
}
