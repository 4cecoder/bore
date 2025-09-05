package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	connectionsTotal  int64
	bytesTransferred  int64
	activeConnections int64
)

type ConnectionPool struct {
	targetAddr string
	pool       chan net.Conn
	mu         sync.Mutex
	maxSize    int
}

func NewConnectionPool(targetAddr string, maxSize int) *ConnectionPool {
	return &ConnectionPool{
		targetAddr: targetAddr,
		pool:       make(chan net.Conn, maxSize),
		maxSize:    maxSize,
	}
}

func (cp *ConnectionPool) Get() (net.Conn, error) {
	select {
	case conn := <-cp.pool:
		// Test if connection is still alive
		conn.SetReadDeadline(time.Now().Add(1 * time.Millisecond))
		var buf [1]byte
		if _, err := conn.Read(buf[:]); err == nil {
			conn.SetReadDeadline(time.Time{})
			return conn, nil
		}
		conn.Close()
		// Fall through to create new connection
	default:
	}

	// Create new connection
	conn, err := net.Dial("tcp", cp.targetAddr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (cp *ConnectionPool) Put(conn net.Conn) {
	select {
	case cp.pool <- conn:
		// Successfully returned to pool
	default:
		// Pool is full, close connection
		conn.Close()
	}
}

func main() {
	var listenPort = flag.Int("port", 8080, "Port to listen on")
	var targetAddr = flag.String("target", "localhost:3000", "Target address to forward to")
	var expectedApiKey = flag.String("api-key", "default-key", "Expected API key for authentication")
	var healthPort = flag.Int("health-port", 8081, "Port for health check endpoint")
	flag.Parse()

	if *listenPort < 1 || *listenPort > 65535 {
		log.Fatal("Invalid port")
	}

	fmt.Printf("Starting bore server on port %d\n", *listenPort)

	// Initialize connection pool
	pool := NewConnectionPool(*targetAddr, 10) // Pool size of 10

	// Start metrics logging
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			structuredLog("INFO", "Metrics update", map[string]interface{}{
				"event":              "metrics",
				"connections_total":  atomic.LoadInt64(&connectionsTotal),
				"active_connections": atomic.LoadInt64(&activeConnections),
				"bytes_transferred":  atomic.LoadInt64(&bytesTransferred),
			})
		}
	}()

	// Start health check server
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			health := map[string]interface{}{
				"status":             "healthy",
				"timestamp":          time.Now().Format(time.RFC3339),
				"connections_total":  atomic.LoadInt64(&connectionsTotal),
				"active_connections": atomic.LoadInt64(&activeConnections),
				"bytes_transferred":  atomic.LoadInt64(&bytesTransferred),
			}
			json.NewEncoder(w).Encode(health)
		})
		http.ListenAndServe(fmt.Sprintf(":%d", *healthPort), nil)
	}()

	cert, err := tls.LoadX509KeyPair("certs/cert.pem", "certs/key.pem")
	if err != nil {
		log.Fatal("Failed to load TLS cert:", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", *listenPort), tlsConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn, pool, *expectedApiKey)
	}
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	ClientIP  string `json:"client_ip,omitempty"`
	Target    string `json:"target,omitempty"`
	Event     string `json:"event,omitempty"`
}

type countingWriter struct {
	conn net.Conn
}

func (cw *countingWriter) Write(p []byte) (n int, err error) {
	return cw.conn.Write(p)
}

func structuredLog(level, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
	}

	if clientIP, ok := fields["client_ip"].(string); ok {
		entry.ClientIP = clientIP
	}
	if target, ok := fields["target"].(string); ok {
		entry.Target = target
	}
	if event, ok := fields["event"].(string); ok {
		entry.Event = event
	}

	if jsonData, err := json.Marshal(entry); err == nil {
		fmt.Println(string(jsonData))
	} else {
		log.Printf("[%s] %s", level, message)
	}
}

func handleConnection(conn net.Conn, pool *ConnectionPool, expectedApiKey string) {
	defer conn.Close()
	atomic.AddInt64(&activeConnections, 1)
	atomic.AddInt64(&connectionsTotal, 1)
	defer atomic.AddInt64(&activeConnections, -1)

	// Read API key
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		structuredLog("ERROR", "Failed to read API key", map[string]interface{}{
			"client_ip": conn.RemoteAddr().String(),
			"event":     "auth_failed",
		})
		return
	}
	apiKey := strings.TrimSpace(scanner.Text())
	if apiKey != expectedApiKey {
		structuredLog("WARN", "Invalid API key", map[string]interface{}{
			"client_ip": conn.RemoteAddr().String(),
			"event":     "auth_failed",
		})
		return
	}

	// Get connection from pool
	targetConn, err := pool.Get()
	if err != nil {
		structuredLog("ERROR", "Failed to get connection from pool", map[string]interface{}{
			"client_ip": conn.RemoteAddr().String(),
			"target":    pool.targetAddr,
			"event":     "connection_failed",
			"error":     err.Error(),
		})
		return
	}
	defer pool.Put(targetConn)

	structuredLog("INFO", "Starting tunnel", map[string]interface{}{
		"client_ip": conn.RemoteAddr().String(),
		"target":    pool.targetAddr,
		"event":     "tunnel_started",
	})

	// Forward data in both directions with metrics
	go func() {
		bytes, _ := io.Copy(targetConn, conn)
		atomic.AddInt64(&bytesTransferred, bytes)
	}()
	bytes, _ := io.Copy(conn, targetConn)
	atomic.AddInt64(&bytesTransferred, bytes)
}
