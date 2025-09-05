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

## Installation

```bash
go install github.com/4cecoder/bore@latest
```

## Usage

### Server
```bash
bore server --port 8080
```

### Client
```bash
bore client --server example.com --local 3000
```

## Documentation

See [PRD.md](PRD.md) for detailed product requirements and epics.

## Contributing

Check the [GitHub issues](https://github.com/4cecoder/bore/issues) for current epics and tasks.

## License

MIT License - see [LICENSE](LICENSE) for details.