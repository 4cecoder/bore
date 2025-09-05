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
	"sync/atomic"
	"time"
)

var (
	connectionsTotal  int64
	bytesTransferred  int64
	activeConnections int64
)

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
		go handleConnection(conn, *targetAddr, *expectedApiKey)
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

func handleConnection(conn net.Conn, targetAddr string, expectedApiKey string) {
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

	// Connect to target
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		structuredLog("ERROR", "Failed to connect to target", map[string]interface{}{
			"client_ip": conn.RemoteAddr().String(),
			"target":    targetAddr,
			"event":     "connection_failed",
			"error":     err.Error(),
		})
		return
	}
	defer targetConn.Close()

	structuredLog("INFO", "Starting tunnel", map[string]interface{}{
		"client_ip": conn.RemoteAddr().String(),
		"target":    targetAddr,
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
