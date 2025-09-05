# bore

A simple, enterprise-grade ngrok alternative written in Golang.

## Overview

bore provides secure tunneling capabilities to expose local development servers to the internet, with a focus on simplicity, security, and scalability.

## Features

- Bidirectional TCP/UDP tunneling
- HTTPS support with automatic TLS
- Authentication and authorization
- Real-time monitoring and logging
- Custom domain support
- Docker containerization
- CI/CD pipelines
- Cross-platform binary releases
- Cloud deployment support

## Installation

### From Source
```bash
git clone https://github.com/4cecoder/bore.git
cd bore
go build -o bin/server ./cmd/server
go build -o bin/client ./cmd/client
```

### Using Docker
```bash
# Build images
docker build --target server -t bore-server .
docker build --target client -t bore-client .

# Or use docker-compose for development
docker-compose up
```

### Pre-built Binaries
Download the latest release from [GitHub Releases](https://github.com/4cecoder/bore/releases).

## Usage

### Server
```bash
./bin/server
```

### Client
```bash
./bin/client -local-port 3000 -server localhost:8080 -api-key your-api-key
```

## Deployment

### Local Development
```bash
docker-compose up
```

### Cloud Deployment (AWS)
```bash
cd terraform/aws
terraform init
terraform plan
terraform apply
```

### Configuration
Environment-specific configurations are available in the `config/` directory:
- `config/dev.yaml` - Development environment
- `config/staging.yaml` - Staging environment
- `config/prod.yaml` - Production environment

## Development

### Running Tests
```bash
go test ./...
```

### Building for Multiple Platforms
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o bin/bore-server-linux-amd64 ./cmd/server
GOOS=linux GOARCH=arm64 go build -o bin/bore-server-linux-arm64 ./cmd/server

# macOS
GOOS=darwin GOARCH=amd64 go build -o bin/bore-server-darwin-amd64 ./cmd/server
GOOS=darwin GOARCH=arm64 go build -o bin/bore-server-darwin-arm64 ./cmd/server

# Windows
GOOS=windows GOARCH=amd64 go build -o bin/bore-server-windows-amd64.exe ./cmd/server
```

## Backup and Restore

### Creating Backups
```bash
./scripts/backup.sh
```

### Restoring from Backup
```bash
./scripts/restore.sh bore_backup_20231201_120000.tar.gz
```

## Documentation

See [PRD.md](PRD.md) for detailed product requirements and epics.

## Contributing

Check the [GitHub issues](https://github.com/4cecoder/bore/issues) for current epics and tasks.

## License

MIT License - see [LICENSE](LICENSE) for details.