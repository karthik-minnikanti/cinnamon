# Network Monitor

A network monitoring tool that tracks connection errors and records relevant information. The tool consists of two main components: a collector agent and a server.

## Architecture

### Collector Agent
- Runs on the target machine
- Monitors network connections
- Detects connection errors
- Sends data to the server

### Server
- Receives and stores connection data
- Provides a web interface
- Offers a REST API for data access

## Features
- Real-time network connection monitoring
- Error detection and tracking
- Process information collection
- Web-based dashboard
- REST API for data access

## Prerequisites
- Go 1.16 or later
- SQLite3
- Network access between collector and server

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/network-monitor.git
cd network-monitor
```

2. Install dependencies:
```bash
go mod download
```

## Building

Build both components:

```bash
# Build the collector
go build -o bin/collector cmd/collector/main.go

# Build the server
go build -o bin/server cmd/server/main.go
```

## Running

### Start the Server
```bash
# Start the server on port 8080
./bin/server -port 8080

# Or with custom database path
./bin/server -port 8080 -db /path/to/network.db
```

### Start the Collector
```bash
# Start the collector with default settings
./bin/collector

# Or with custom server URL and interval
./bin/collector -server http://localhost:8080 -interval 2s
```

### Access the Web Interface
Open your browser and navigate to:
```
http://localhost:8080
```

## API Endpoints

### Submit Connection Data
```
POST /api/connections
Content-Type: application/json

{
    "timestamp": "2024-03-16T12:00:00Z",
    "process_id": 1234,
    "process_name": "example",
    "process_path": "/usr/bin/example",
    "local_addr": "127.0.0.1:8080",
    "remote_addr": "192.168.1.1:12345",
    "error": "ECONNREFUSED",
    "protocol": "tcp",
    "state": "LISTEN"
}
```

### Get Connections
```
GET /api/connections
```

### Filter Connections by Error
```
GET /api/connections?error=ECONNREFUSED
```

## Database Schema

The tool uses SQLite to store connection data with the following schema:

```sql
CREATE TABLE connections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL,
    process_id INTEGER,
    process_name TEXT,
    process_path TEXT,
    local_addr TEXT NOT NULL,
    remote_addr TEXT NOT NULL,
    error TEXT,
    protocol TEXT,
    state TEXT
);
```

## Error Types Tracked
- ECONNREFUSED: Connection refused
- ECONNABORTED: Connection aborted
- ECONNRESET: Connection reset
- ETIMEDOUT: Connection timeout

## Contributing
1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License
This project is licensed under the MIT License - see the LICENSE file for details. 