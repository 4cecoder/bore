package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

func main() {
	var localPort = flag.Int("local-port", 8080, "Local port to tunnel")
	var serverAddr = flag.String("server", "localhost:8080", "Server address")
	var apiKey = flag.String("api-key", "", "API key for authentication")
	flag.Parse()

	if *localPort < 1 || *localPort > 65535 {
		log.Fatal("Invalid local port")
	}

	fmt.Printf("Starting bore client: tunneling local port %d to %s\n", *localPort, *serverAddr)

	// Listen on local port
	localListener, err := net.Listen("tcp", fmt.Sprintf(":%d", *localPort))
	if err != nil {
		log.Fatal("Failed to listen on local port:", err)
	}
	defer localListener.Close()

	fmt.Println("Client ready. Waiting for connections...")

	for {
		localConn, err := localListener.Accept()
		if err != nil {
			log.Println("Error accepting local connection:", err)
			continue
		}
		go handleLocalConnection(localConn, *serverAddr, *apiKey)
	}
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	LocalAddr string `json:"local_addr,omitempty"`
	Server    string `json:"server,omitempty"`
	Event     string `json:"event,omitempty"`
}

func structuredLog(level, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
	}

	if localAddr, ok := fields["local_addr"].(string); ok {
		entry.LocalAddr = localAddr
	}
	if server, ok := fields["server"].(string); ok {
		entry.Server = server
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

func handleLocalConnection(localConn net.Conn, serverAddr string, apiKey string) {
	defer localConn.Close()

	// Connect to server for this connection
	serverConn, err := tls.Dial("tcp", serverAddr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		structuredLog("ERROR", "Failed to connect to server", map[string]interface{}{
			"local_addr": localConn.RemoteAddr().String(),
			"server":     serverAddr,
			"event":      "connection_failed",
			"error":      err.Error(),
		})
		return
	}
	defer serverConn.Close()

	// Send API key for authentication
	if _, err := fmt.Fprintf(serverConn, "%s\n", apiKey); err != nil {
		structuredLog("ERROR", "Failed to send API key", map[string]interface{}{
			"local_addr": localConn.RemoteAddr().String(),
			"server":     serverAddr,
			"event":      "auth_failed",
			"error":      err.Error(),
		})
		return
	}

	structuredLog("INFO", "Starting tunnel", map[string]interface{}{
		"local_addr": localConn.RemoteAddr().String(),
		"server":     serverAddr,
		"event":      "tunnel_started",
	})

	// Forward data in both directions
	go io.Copy(serverConn, localConn)
	io.Copy(localConn, serverConn)
}
