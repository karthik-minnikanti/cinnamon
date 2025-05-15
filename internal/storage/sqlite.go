package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/karthik-minnikanti/cinnamon/internal/models"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteStorage{db: db}, nil
}

func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS connections (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		process_id INTEGER NOT NULL,
		process_name TEXT NOT NULL,
		process_path TEXT NOT NULL,
		local_addr TEXT NOT NULL,
		remote_addr TEXT NOT NULL,
		error TEXT,
		service_name TEXT,
		protocol TEXT NOT NULL,
		state TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_connections_timestamp ON connections(timestamp);
	CREATE INDEX IF NOT EXISTS idx_connections_process_id ON connections(process_id);
	CREATE INDEX IF NOT EXISTS idx_connections_error ON connections(error);
	`

	_, err := db.Exec(query)
	return err
}

func (s *SQLiteStorage) StoreConnection(conn *models.Connection) error {
	query := `
	INSERT INTO connections (
		timestamp, process_id, process_name, process_path,
		local_addr, remote_addr, error, service_name,
		protocol, state
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		conn.Timestamp,
		conn.ProcessID,
		conn.ProcessName,
		conn.ProcessPath,
		conn.LocalAddr,
		conn.RemoteAddr,
		conn.Error,
		conn.ServiceName,
		conn.Protocol,
		conn.State,
	)
	return err
}

func (s *SQLiteStorage) GetConnections(limit, offset int) ([]*models.Connection, error) {
	query := `
	SELECT id, timestamp, process_id, process_name, process_path,
		local_addr, remote_addr, error, service_name, protocol, state
	FROM connections
	ORDER BY timestamp DESC
	LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []*models.Connection
	for rows.Next() {
		conn := &models.Connection{}
		var timestamp string
		err := rows.Scan(
			&conn.ID,
			&timestamp,
			&conn.ProcessID,
			&conn.ProcessName,
			&conn.ProcessPath,
			&conn.LocalAddr,
			&conn.RemoteAddr,
			&conn.Error,
			&conn.ServiceName,
			&conn.Protocol,
			&conn.State,
		)
		if err != nil {
			return nil, err
		}
		conn.Timestamp, _ = time.Parse(time.RFC3339, timestamp)
		connections = append(connections, conn)
	}
	return connections, nil
}

func (s *SQLiteStorage) GetConnectionsByError(error models.ConnectionError) ([]*models.Connection, error) {
	query := `
	SELECT id, timestamp, process_id, process_name, process_path,
		local_addr, remote_addr, error, service_name, protocol, state
	FROM connections
	WHERE error = ?
	ORDER BY timestamp DESC
	`

	rows, err := s.db.Query(query, error)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []*models.Connection
	for rows.Next() {
		conn := &models.Connection{}
		var timestamp string
		err := rows.Scan(
			&conn.ID,
			&timestamp,
			&conn.ProcessID,
			&conn.ProcessName,
			&conn.ProcessPath,
			&conn.LocalAddr,
			&conn.RemoteAddr,
			&conn.Error,
			&conn.ServiceName,
			&conn.Protocol,
			&conn.State,
		)
		if err != nil {
			return nil, err
		}
		conn.Timestamp, _ = time.Parse(time.RFC3339, timestamp)
		connections = append(connections, conn)
	}
	return connections, nil
}

func (s *SQLiteStorage) GetConnectionsByProcess(processID int) ([]*models.Connection, error) {
	query := `
	SELECT id, timestamp, process_id, process_name, process_path,
		local_addr, remote_addr, error, service_name, protocol, state
	FROM connections
	WHERE process_id = ?
	ORDER BY timestamp DESC
	`

	rows, err := s.db.Query(query, processID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []*models.Connection
	for rows.Next() {
		conn := &models.Connection{}
		var timestamp string
		err := rows.Scan(
			&conn.ID,
			&timestamp,
			&conn.ProcessID,
			&conn.ProcessName,
			&conn.ProcessPath,
			&conn.LocalAddr,
			&conn.RemoteAddr,
			&conn.Error,
			&conn.ServiceName,
			&conn.Protocol,
			&conn.State,
		)
		if err != nil {
			return nil, err
		}
		conn.Timestamp, _ = time.Parse(time.RFC3339, timestamp)
		connections = append(connections, conn)
	}
	return connections, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
