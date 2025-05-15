# Network Connection Monitor

A Go-based network monitoring tool that tracks network connections, connection errors, and process information. The tool consists of two components: a collector agent that gathers network data and a server that stores and displays the information.

## Architecture

The tool is split into two main components:

1. **Collector Agent**
   - Runs on the target machine
   - Monitors network connections in real-time
   - Collects process information
   - Sends data to the server

2. **Server**
   - Receives and stores connection data
   - Provides a web interface for monitoring
   - Offers REST API for data access
   - Stores data in SQLite database

## Features

- Real-time network connection monitoring
- Connection error tracking (ECONNREFUSED, ECONNABORTED, etc.)
- Process information collection (PID, process name, path)
- Web-based dashboard for monitoring
- REST API for data access
- Cross-platform support (Linux, macOS, Windows)

## Prerequisites

- Go 1.21 or later
- SQLite3
- System tools:
  - netstat
  - lsof (on Linux/macOS)

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

3. Build both components:
```bash
# Build collector
go build -o network-monitor-collector ./cmd/collector/main.go

# Build server
go build -o network-monitor-server ./cmd/server/main.go
```

## Usage

1. Start the server:
```bash
./network-monitor-server -port 8080 -db network_monitor.db
```

2. Start the collector on the target machine:
```bash
./network-monitor-collector -server http://localhost:8080
```

3. Access the web interface at `http://localhost:8080`

## API Endpoints

- `POST /api/connections` - Submit new connection data
- `GET /api/connections` - Get all connections (with optional limit/offset)
- `GET /api/connections/error/{error}` - Get connections with specific error
- `GET /api/connections/process/{pid}` - Get connections for specific process

## Database Schema

The tool uses SQLite to store connection information with the following schema:

```sql
CREATE TABLE connections (
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
```

## Error Types Tracked

- ECONNREFUSED: Connection refused
- ECONNABORTED: Connection aborted
- ECONNRESET: Connection reset
- ECONNTIMEOUT: Connection timeout
- EHOSTUNREACH: Host unreachable
- ENETUNREACH: Network unreachable

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 