# Multi-stage Dockerfile for bore (client and server)

# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/

# Build server binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Build client binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o client ./cmd/client

# Final stage for server
FROM alpine:latest AS server

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy server binary and certificates
COPY --from=builder /app/server .
COPY certs/ ./certs/

EXPOSE 8080 8443

CMD ["./server"]

# Final stage for client
FROM alpine:latest AS client

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy client binary
COPY --from=builder /app/client .

CMD ["./client"]