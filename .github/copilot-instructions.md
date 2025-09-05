# bore - Enterprise-grade ngrok Alternative

bore is a secure tunneling solution written in Go that provides bidirectional TCP/UDP tunneling, HTTPS support with automatic TLS, authentication, real-time monitoring, and cloud deployment capabilities.

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Bootstrap, Build, and Test the Repository
- Check Go version: `go version` (requires Go 1.25+)
- Download dependencies: `go mod download` -- takes 1-2 seconds, no dependencies currently
- Create build directory: `mkdir -p bin`
- **Build server**: `go build -o bin/server ./cmd/server` -- takes 10-15 seconds. NEVER CANCEL. Set timeout to 60+ seconds.
- **Build client**: `go build -o bin/client ./cmd/client` -- takes 1-2 seconds. Set timeout to 30+ seconds.
- **Run tests**: `go test -v ./...` -- takes 2-3 seconds. NEVER CANCEL. Set timeout to 30+ seconds.
- **Run go vet**: `go vet ./...` -- takes 1-2 seconds. Set timeout to 30+ seconds.

### Cross-platform Builds
- **Linux AMD64**: `GOOS=linux GOARCH=amd64 go build -o bin/bore-server-linux-amd64 ./cmd/server` -- takes 1-2 seconds
- **Linux ARM64**: `GOOS=linux GOARCH=arm64 go build -o bin/bore-server-linux-arm64 ./cmd/server` -- takes 1-2 seconds  
- **macOS AMD64**: `GOOS=darwin GOARCH=amd64 go build -o bin/bore-server-darwin-amd64 ./cmd/server` -- takes 8-12 seconds. NEVER CANCEL. Set timeout to 60+ seconds.
- **macOS ARM64**: `GOOS=darwin GOARCH=arm64 go build -o bin/bore-server-darwin-arm64 ./cmd/server` -- takes 8-12 seconds. NEVER CANCEL. Set timeout to 60+ seconds.
- **Windows AMD64**: `GOOS=windows GOARCH=amd64 go build -o bin/bore-server-windows-amd64.exe ./cmd/server` -- takes 8-12 seconds. NEVER CANCEL. Set timeout to 60+ seconds.

### TLS Certificates
- **CRITICAL**: TLS certificates are REQUIRED and already exist in `certs/cert.pem` and `certs/key.pem`
- Server will fail to start without valid certificates
- Certificates are self-signed and valid for localhost development
- Client uses `InsecureSkipVerify: true` for testing with self-signed certificates

### Running the Application

#### Server
- **Basic server**: `./bin/server -port 8080 -target localhost:3000 -api-key test-key`
- **Health check endpoint**: Available at `http://localhost:8081/health` (port 8081)
- **Default ports**: Server listens on 8080 (TLS), health check on 8081
- **Supported flags**:
  - `-port`: Server port (default: 8080)
  - `-target`: Target address to forward to (default: localhost:3000) 
  - `-api-key`: Expected API key for authentication (default: default-key)
  - `-health-port`: Health check port (default: 8081)
  - `-max-connections`: Maximum concurrent connections (default: 100)

#### Client  
- **Basic client**: `./bin/client -local-port 3000 -server localhost:8080 -api-key test-key`
- **Supported flags**:
  - `-local-port`: Local port to tunnel (default: 8080)
  - `-server`: Server address (default: localhost:8080)
  - `-api-key`: API key for authentication (required for connection)

### Docker Build and Deployment
- **Docker build fails** due to network restrictions in many environments. Do not rely on Docker builds working.
- **If Docker build works**: 
  - `docker build --target server -t bore-server .` -- takes 2-5 minutes. NEVER CANCEL. Set timeout to 10+ minutes.
  - `docker build --target client -t bore-client .` -- takes 2-5 minutes. NEVER CANCEL. Set timeout to 10+ minutes.
- **Docker Compose**: `docker compose up -d` (modern syntax, not `docker-compose`)
  - `docker compose ps` to check status
  - `docker compose down` to stop services

### Backup and Restore Operations
- **Create backup**: `./scripts/backup.sh` -- creates timestamped backup in `/opt/bore/backups/`
- **Requires setup**: `sudo mkdir -p /opt/bore/backups && sudo chown -R $(whoami):$(whoami) /opt/bore`
- **Restore backup**: `./scripts/restore.sh bore_backup_YYYYMMDD_HHMMSS.tar.gz`
- **Backup includes**: Configuration files, certificates, logs (if present)
- **Backup time**: Takes 2-5 seconds, creates ~8KB archive

## Validation

### Manual Testing Scenarios
- **ALWAYS run through complete end-to-end scenarios** after making changes:
  1. Start a simple web server: `python3 -m http.server 3000`
  2. Start bore server: `./bin/server -port 8080 -target localhost:3000 -api-key test-key`
  3. Check health endpoint: `curl http://localhost:8081/health`
  4. Test client connection (advanced scenario)
  5. Verify structured JSON logging output

### Health Check Validation
- **Health endpoint**: `curl http://localhost:8081/health`
- **Expected response**: JSON with status, timestamp, connections metrics
- **Example**: `{"active_connections":0,"bytes_transferred":0,"connections_total":0,"status":"healthy","timestamp":"2025-09-05T19:23:26Z"}`

### Build Verification
- **Verify binaries**: `ls -la bin/ && file bin/server bin/client`
- **Test help output**: `./bin/server -h` and `./bin/client -h`
- **Size expectations**: Server ~9MB, Client ~7MB (Linux AMD64)

## Configuration

### Environment-specific Configurations
Configuration files are available in the `config/` directory:
- `config/dev.yaml` - Development environment with debug logging
- `config/staging.yaml` - Staging environment 
- `config/prod.yaml` - Production environment

### Development Configuration Highlights
- Server port: 8080, TLS port: 8443
- API key: "dev-api-key-12345"
- Debug logging enabled
- Metrics on port 9090
- Max connections: 100

## CI/CD and Quality

### GitHub Actions Workflows
- **CI**: `.github/workflows/ci.yml` - builds, tests, linting, Docker builds
- **Integration**: `.github/workflows/integration.yml` - end-to-end tests with Redis
- **Release**: `.github/workflows/release.yml` - automated releases
- **Security**: `.github/workflows/security.yml` - security scanning

### Linting and Code Quality
- **go vet works**: `go vet ./...` -- takes 1-2 seconds
- **golangci-lint has version issues** with Go 1.25. Skip golangci-lint or use older Go version.
- **Commit message format**: Must follow conventional commits (e.g., "feat: add new feature")

### Pre-commit Validation
Before committing changes, ALWAYS run:
1. `go mod download`
2. `mkdir -p bin`
3. `go build -o bin/server ./cmd/server`
4. `go build -o bin/client ./cmd/client` 
5. `go test -v ./...`
6. `go vet ./...`
7. Test application functionality as described in Validation section

## Common Tasks

### Repository Structure
```
bore/
├── .github/workflows/     # CI/CD pipelines
├── cmd/
│   ├── client/           # Client application code
│   └── server/           # Server application code  
├── config/               # Environment configurations
├── certs/               # TLS certificates (cert.pem, key.pem)
├── scripts/             # Operational scripts (backup.sh, restore.sh)
├── terraform/           # Infrastructure as code
├── go.mod              # Go module definition
├── Dockerfile          # Multi-stage Docker build
├── docker-compose.yml  # Development environment
└── README.md           # Project documentation
```

### Key Files to Monitor
- **After changing server code**: Always rebuild and test `./bin/server`
- **After changing client code**: Always rebuild and test `./bin/client`
- **After changing configs**: Restart services and test with new configuration
- **After changing certificates**: Restart server, test TLS connections

### Debugging Tips
- **Server logs**: Structured JSON output to stdout
- **Health endpoint**: Real-time metrics and status
- **Connection issues**: Check API key authentication
- **TLS issues**: Verify certificates exist and are valid
- **Port conflicts**: Check if ports 8080, 8081 are available

### Infrastructure and Deployment
- **Terraform**: Available in `terraform/` directory for cloud deployment
- **Multiple targets**: Server supports comma-separated target addresses for load balancing
- **Connection pooling**: Built-in connection pool with configurable max size
- **Metrics**: Real-time connection and traffic metrics available

### Performance Characteristics
- **Lightweight**: ~16MB total binary size for both server and client
- **Fast builds**: Server builds in 10-15 seconds, client in 1-2 seconds
- **Quick tests**: Full test suite runs in 2-3 seconds
- **Low resource usage**: Suitable for containerized environments
- **Connection limits**: Configurable max connections (default: 100)

### Troubleshooting Common Issues
- **"Failed to load TLS cert"**: Ensure `certs/cert.pem` and `certs/key.pem` exist
- **"Invalid API key"**: Check API key matches between server and client
- **Connection refused**: Verify target service is running on specified port
- **Docker build failures**: Network restrictions may prevent package downloads
- **Go version issues**: Requires Go 1.25+, some tools may have compatibility issues