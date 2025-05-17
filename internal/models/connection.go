package models

import (
	"time"
)

// ConnectionError represents different types of connection errors
type ConnectionError string

const (
	ErrConnRefused    ConnectionError = "ECONNREFUSED"
	ErrConnAborted    ConnectionError = "ECONNABORTED"
	ErrConnReset      ConnectionError = "ECONNRESET"
	ErrConnTimeout    ConnectionError = "ETIMEDOUT"
	ErrDNSFailure     ConnectionError = "EDNSFAILURE"
	ErrHostUnreach    ConnectionError = "EHOSTUNREACH"
	ErrNetworkDown    ConnectionError = "ENETDOWN"
	ErrNetworkUnreach ConnectionError = "ENETUNREACH"
)

// ServiceType represents the type of service
type ServiceType string

const (
	ServiceTypeDatabase     ServiceType = "database"
	ServiceTypeMessageQueue ServiceType = "message_queue"
	ServiceTypeCache        ServiceType = "cache"
	ServiceTypeAPI          ServiceType = "api"
	ServiceTypeOther        ServiceType = "other"
)

// DatabaseType represents specific database types
type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypeMongoDB    DatabaseType = "mongodb"
	DatabaseTypeRedis      DatabaseType = "redis"
	DatabaseTypeOther      DatabaseType = "other"
)

// MessageQueueType represents specific message queue types
type MessageQueueType string

const (
	MessageQueueTypeRabbitMQ MessageQueueType = "rabbitmq"
	MessageQueueTypeKafka    MessageQueueType = "kafka"
	MessageQueueTypeOther    MessageQueueType = "other"
)

// Connection represents a network connection with extended metadata
type Connection struct {
	ID               string                 `json:"id"`
	Timestamp        time.Time              `json:"timestamp"`
	SourceIP         string                 `json:"source_ip"`
	SourcePort       int                    `json:"source_port"`
	DestIP           string                 `json:"dest_ip"`
	DestPort         int                    `json:"dest_port"`
	Protocol         string                 `json:"protocol"`
	ServiceName      string                 `json:"service_name"`
	ServiceType      ServiceType            `json:"service_type"`
	DatabaseType     DatabaseType           `json:"database_type,omitempty"`
	MessageQueueType MessageQueueType       `json:"message_queue_type,omitempty"`
	Host             string                 `json:"host"`
	DeploymentID     string                 `json:"deployment_id"`
	Environment      string                 `json:"environment"`
	Region           string                 `json:"region"`
	Latency          float64                `json:"latency_ms"`
	BytesSent        int64                  `json:"bytes_sent"`
	BytesReceived    int64                  `json:"bytes_received"`
	RetryCount       int                    `json:"retry_count"`
	Error            string                 `json:"error,omitempty"`
	Tags             []string               `json:"tags"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// ConnectionStats represents aggregated statistics
type ConnectionStats struct {
	TotalConnections   int64                    `json:"total_connections"`
	ErrorCounts        map[string]int64         `json:"error_counts"`
	AvgLatency         float64                  `json:"avg_latency"`
	TotalBytesSent     int64                    `json:"total_bytes_sent"`
	TotalBytesReceived int64                    `json:"total_bytes_received"`
	TopServices        []ServiceStats           `json:"top_services"`
	ErrorTrends        []ErrorTrend             `json:"error_trends"`
	ServiceTypeStats   map[ServiceType]int      `json:"service_type_stats"`
	DatabaseStats      map[DatabaseType]int     `json:"database_stats"`
	QueueStats         map[MessageQueueType]int `json:"queue_stats"`
}

// ServiceStats represents statistics for a service
type ServiceStats struct {
	ServiceName   string  `json:"service_name"`
	ErrorCount    int64   `json:"error_count"`
	AvgLatency    float64 `json:"avg_latency"`
	TotalRequests int64   `json:"total_requests"`
}

// ErrorTrend represents error trends over time
type ErrorTrend struct {
	Timestamp time.Time `json:"timestamp"`
	ErrorType string    `json:"error_type"`
	Count     int64     `json:"count"`
}
