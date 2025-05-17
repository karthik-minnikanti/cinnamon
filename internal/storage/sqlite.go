package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/karthik-minnikanti/cinnamon/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	storage := &SQLiteStorage{db: db}
	if err := storage.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	return storage, nil
}

func (s *SQLiteStorage) migrate() error {
	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Drop existing table if it exists
	_, err = tx.Exec("DROP TABLE IF EXISTS connections")
	if err != nil {
		return fmt.Errorf("failed to drop existing table: %v", err)
	}

	// Create the table with all required columns
	_, err = tx.Exec(`
		CREATE TABLE connections (
			id TEXT PRIMARY KEY,
			timestamp DATETIME NOT NULL,
			source_ip TEXT NOT NULL,
			source_port INTEGER NOT NULL,
			dest_ip TEXT NOT NULL,
			dest_port INTEGER NOT NULL,
			protocol TEXT NOT NULL,
			service_name TEXT,
			service_type TEXT,
			database_type TEXT,
			message_queue_type TEXT,
			host TEXT,
			deployment_id TEXT,
			environment TEXT,
			region TEXT,
			latency_ms REAL,
			bytes_sent INTEGER,
			bytes_received INTEGER,
			retry_count INTEGER,
			error TEXT,
			tags TEXT,
			metadata TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX idx_connections_timestamp ON connections(timestamp)",
		"CREATE INDEX idx_connections_service ON connections(service_name)",
		"CREATE INDEX idx_connections_error ON connections(error)",
		"CREATE INDEX idx_connections_environment ON connections(environment)",
		"CREATE INDEX idx_connections_service_type ON connections(service_type)",
	}

	for _, idx := range indexes {
		if _, err := tx.Exec(idx); err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (s *SQLiteStorage) StoreConnection(conn *models.Connection) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	tags, err := json.Marshal(conn.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %v", err)
	}

	metadata, err := json.Marshal(conn.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO connections (
			id, timestamp, source_ip, source_port, dest_ip, dest_port,
			protocol, service_name, service_type, database_type, message_queue_type,
			host, deployment_id, environment, region, latency_ms,
			bytes_sent, bytes_received, retry_count, error, tags, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		conn.ID, conn.Timestamp, conn.SourceIP, conn.SourcePort, conn.DestIP, conn.DestPort,
		conn.Protocol, conn.ServiceName, conn.ServiceType, conn.DatabaseType, conn.MessageQueueType,
		conn.Host, conn.DeploymentID, conn.Environment, conn.Region, conn.Latency,
		conn.BytesSent, conn.BytesReceived, conn.RetryCount, conn.Error, tags, metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to store connection: %v", err)
	}

	return nil
}

func (s *SQLiteStorage) GetConnections(service, errorType, environment, search string) ([]*models.Connection, error) {
	query := `
		SELECT id, timestamp, source_ip, source_port, dest_ip, dest_port,
			protocol, service_name, service_type, database_type, message_queue_type,
			host, deployment_id, environment, region, latency_ms,
			bytes_sent, bytes_received, retry_count, error, tags, metadata
		FROM connections
		WHERE 1=1
	`
	args := []interface{}{}

	if service != "" {
		query += " AND service_name = ?"
		args = append(args, service)
	}
	if errorType != "" {
		query += " AND error = ?"
		args = append(args, errorType)
	}
	if environment != "" {
		query += " AND environment = ?"
		args = append(args, environment)
	}
	if search != "" {
		query += " AND (service_name LIKE ? OR host LIKE ? OR deployment_id LIKE ?)"
		searchArg := "%" + search + "%"
		args = append(args, searchArg, searchArg, searchArg)
	}

	query += " ORDER BY timestamp DESC LIMIT 1000"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query connections: %v", err)
	}
	defer rows.Close()

	var connections []*models.Connection
	for rows.Next() {
		var conn models.Connection
		var tags, metadata []byte
		var serviceType, dbType, queueType sql.NullString

		err := rows.Scan(
			&conn.ID, &conn.Timestamp, &conn.SourceIP, &conn.SourcePort, &conn.DestIP, &conn.DestPort,
			&conn.Protocol, &conn.ServiceName, &serviceType, &dbType, &queueType,
			&conn.Host, &conn.DeploymentID, &conn.Environment, &conn.Region, &conn.Latency,
			&conn.BytesSent, &conn.BytesReceived, &conn.RetryCount, &conn.Error, &tags, &metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection: %v", err)
		}

		// Handle NULL values
		if serviceType.Valid {
			conn.ServiceType = models.ServiceType(serviceType.String)
		}
		if dbType.Valid {
			conn.DatabaseType = models.DatabaseType(dbType.String)
		}
		if queueType.Valid {
			conn.MessageQueueType = models.MessageQueueType(queueType.String)
		}

		if err := json.Unmarshal(tags, &conn.Tags); err != nil {
			log.Printf("Warning: failed to unmarshal tags: %v", err)
		}
		if err := json.Unmarshal(metadata, &conn.Metadata); err != nil {
			log.Printf("Warning: failed to unmarshal metadata: %v", err)
		}

		connections = append(connections, &conn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating connections: %v", err)
	}

	return connections, nil
}

func (s *SQLiteStorage) GetConnectionByID(id string) (*models.Connection, error) {
	var conn models.Connection
	var tags, metadata []byte
	err := s.db.QueryRow(`
		SELECT id, timestamp, source_ip, source_port, dest_ip, dest_port,
			protocol, service_name, service_type, database_type, message_queue_type,
			host, deployment_id, environment, region, latency_ms,
			bytes_sent, bytes_received, retry_count, error, tags, metadata
		FROM connections
		WHERE id = ?
	`, id).Scan(
		&conn.ID, &conn.Timestamp, &conn.SourceIP, &conn.SourcePort, &conn.DestIP, &conn.DestPort,
		&conn.Protocol, &conn.ServiceName, &conn.ServiceType, &conn.DatabaseType, &conn.MessageQueueType,
		&conn.Host, &conn.DeploymentID, &conn.Environment, &conn.Region, &conn.Latency,
		&conn.BytesSent, &conn.BytesReceived, &conn.RetryCount, &conn.Error, &tags, &metadata,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("connection not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %v", err)
	}

	if err := json.Unmarshal(tags, &conn.Tags); err != nil {
		log.Printf("Warning: failed to unmarshal tags: %v", err)
	}
	if err := json.Unmarshal(metadata, &conn.Metadata); err != nil {
		log.Printf("Warning: failed to unmarshal metadata: %v", err)
	}

	return &conn, nil
}

func (s *SQLiteStorage) GetStats(startTime, endTime time.Time) (*models.ConnectionStats, error) {
	stats := &models.ConnectionStats{
		ErrorCounts:      make(map[string]int64),
		ServiceTypeStats: make(map[models.ServiceType]int),
		DatabaseStats:    make(map[models.DatabaseType]int),
		QueueStats:       make(map[models.MessageQueueType]int),
	}

	// Get total connections
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM connections
		WHERE timestamp BETWEEN ? AND ?
	`, startTime, endTime).Scan(&stats.TotalConnections)
	if err != nil {
		return nil, fmt.Errorf("failed to get total connections: %v", err)
	}

	// Get error counts
	rows, err := s.db.Query(`
		SELECT error, COUNT(*) FROM connections
		WHERE timestamp BETWEEN ? AND ? AND error IS NOT NULL
		GROUP BY error
	`, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get error counts: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var errorType string
		var count int64
		if err := rows.Scan(&errorType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan error count: %v", err)
		}
		stats.ErrorCounts[errorType] = count
	}

	// Get average latency
	var avgLatency sql.NullFloat64
	err = s.db.QueryRow(`
		SELECT AVG(latency_ms) FROM connections
		WHERE timestamp BETWEEN ? AND ? AND latency_ms IS NOT NULL
	`, startTime, endTime).Scan(&avgLatency)
	if err != nil {
		return nil, fmt.Errorf("failed to get average latency: %v", err)
	}
	if avgLatency.Valid {
		stats.AvgLatency = avgLatency.Float64
	}

	// Get total bytes
	var totalBytesSent, totalBytesReceived sql.NullInt64
	err = s.db.QueryRow(`
		SELECT SUM(bytes_sent), SUM(bytes_received) FROM connections
		WHERE timestamp BETWEEN ? AND ?
	`, startTime, endTime).Scan(&totalBytesSent, &totalBytesReceived)
	if err != nil {
		return nil, fmt.Errorf("failed to get total bytes: %v", err)
	}
	if totalBytesSent.Valid {
		stats.TotalBytesSent = totalBytesSent.Int64
	}
	if totalBytesReceived.Valid {
		stats.TotalBytesReceived = totalBytesReceived.Int64
	}

	// Get service type stats
	rows, err = s.db.Query(`
		SELECT service_type, COUNT(*) FROM connections
		WHERE timestamp BETWEEN ? AND ? AND service_type IS NOT NULL
		GROUP BY service_type
	`, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get service type stats: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var serviceType sql.NullString
		var count int
		if err := rows.Scan(&serviceType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan service type: %v", err)
		}
		if serviceType.Valid {
			stats.ServiceTypeStats[models.ServiceType(serviceType.String)] = count
		}
	}

	// Get database stats
	rows, err = s.db.Query(`
		SELECT database_type, COUNT(*) FROM connections
		WHERE timestamp BETWEEN ? AND ? AND service_type = 'database' AND database_type IS NOT NULL
		GROUP BY database_type
	`, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get database stats: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dbType sql.NullString
		var count int
		if err := rows.Scan(&dbType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan database type: %v", err)
		}
		if dbType.Valid {
			stats.DatabaseStats[models.DatabaseType(dbType.String)] = count
		}
	}

	// Get queue stats
	rows, err = s.db.Query(`
		SELECT message_queue_type, COUNT(*) FROM connections
		WHERE timestamp BETWEEN ? AND ? AND service_type = 'message_queue' AND message_queue_type IS NOT NULL
		GROUP BY message_queue_type
	`, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue stats: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var queueType sql.NullString
		var count int
		if err := rows.Scan(&queueType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan queue type: %v", err)
		}
		if queueType.Valid {
			stats.QueueStats[models.MessageQueueType(queueType.String)] = count
		}
	}

	return stats, nil
}

func (s *SQLiteStorage) GetServices() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT service_name FROM connections WHERE service_name != ''")
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %v", err)
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var service string
		if err := rows.Scan(&service); err != nil {
			return nil, fmt.Errorf("failed to scan service: %v", err)
		}
		services = append(services, service)
	}
	return services, nil
}

func (s *SQLiteStorage) GetErrors() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT error FROM connections WHERE error != ''")
	if err != nil {
		return nil, fmt.Errorf("failed to get errors: %v", err)
	}
	defer rows.Close()

	var errors []string
	for rows.Next() {
		var errorType string
		if err := rows.Scan(&errorType); err != nil {
			return nil, fmt.Errorf("failed to scan error: %v", err)
		}
		errors = append(errors, errorType)
	}
	return errors, nil
}

func (s *SQLiteStorage) GetEnvironments() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT environment FROM connections WHERE environment != ''")
	if err != nil {
		return nil, fmt.Errorf("failed to get environments: %v", err)
	}
	defer rows.Close()

	var environments []string
	for rows.Next() {
		var env string
		if err := rows.Scan(&env); err != nil {
			return nil, fmt.Errorf("failed to scan environment: %v", err)
		}
		environments = append(environments, env)
	}
	return environments, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
