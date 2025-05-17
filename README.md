# Cinnamon

Network Monitoring & Analytics

A modern, minimal network monitoring tool that provides real-time insights into your network connections, services, and potential issues.

## Use Cases

### 1. Service Health Monitoring
- Monitor database connections (PostgreSQL, MySQL, MongoDB, Redis)
- Track message queue health (RabbitMQ, Kafka)
- Watch API endpoint availability
- Detect connection spikes or drops
- Monitor service latency and performance

### 2. Security & Compliance
- Track unauthorized connection attempts
- Monitor connections to sensitive services
- Detect unusual connection patterns
- Log connection metadata for audit trails
- Track connection sources and destinations

### 3. Performance Optimization
- Identify slow connections
- Monitor connection retry patterns
- Track bytes transferred
- Analyze connection duration
- Optimize service configurations

### 4. Troubleshooting
- Real-time error detection
- Connection failure analysis
- Service availability monitoring
- Network latency tracking
- Connection state visualization

### 5. DevOps & Operations
- Monitor microservices communication
- Track service dependencies
- Monitor deployment health
- Track environment-specific issues
- Alert on connection problems

## Example Scenarios

### Database Monitoring
```bash
# Monitor production database connections
go run cmd/collector/main.go \
  --service postgres-prod \
  --host db-prod-1 \
  --deployment v1.2.3 \
  --env production \
  --region us-east
```

### API Service Monitoring
```bash
# Monitor API service connections
go run cmd/collector/main.go \
  --service api-service \
  --host api-prod-1 \
  --deployment api-v2 \
  --env production \
  --region us-west
```

### Message Queue Monitoring
```bash
# Monitor message queue connections
go run cmd/collector/main.go \
  --service kafka-service \
  --host kafka-1 \
  --deployment kafka-v1 \
  --env production \
  --region eu-west
```

## Features
- Real-time connection monitoring
- Service type detection
- Error tracking and analytics
- Beautiful black & white UI
- Detailed connection statistics
- Customizable alerts and notifications

## Getting Started
1. Clone the repository
2. Run the collector: `go run cmd/collector/main.go`
3. Run the server: `go run cmd/server/main.go`
4. Open http://localhost:8080 in your browser

## Configuration
- Collector settings can be configured via command-line flags
- UI settings can be customized through the web interface
- Notification thresholds and retention periods can be adjusted

## License: MIT 