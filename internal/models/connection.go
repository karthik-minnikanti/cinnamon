package models

import (
	"time"
)

// ConnectionError represents different types of network connection errors
type ConnectionError string

const (
	ErrConnRefused    ConnectionError = "ECONNREFUSED"
	ErrConnAborted    ConnectionError = "ECONNABORTED"
	ErrConnReset      ConnectionError = "ECONNRESET"
	ErrConnTimeout    ConnectionError = "ECONNTIMEOUT"
	ErrHostUnreach    ConnectionError = "EHOSTUNREACH"
	ErrNetworkUnreach ConnectionError = "ENETUNREACH"
)

// Connection represents a network connection event
type Connection struct {
	ID          int64           `json:"id"`
	Timestamp   time.Time       `json:"timestamp"`
	ProcessID   int             `json:"process_id"`
	ProcessName string          `json:"process_name"`
	ProcessPath string          `json:"process_path"`
	LocalAddr   string          `json:"local_addr"`
	RemoteAddr  string          `json:"remote_addr"`
	Error       ConnectionError `json:"error,omitempty"`
	ServiceName string          `json:"service_name,omitempty"`
	Protocol    string          `json:"protocol"`
	State       string          `json:"state"`
}
