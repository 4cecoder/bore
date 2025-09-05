package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	var listenPort = flag.Int("port", 8080, "Port to listen on")
	var targetAddr = flag.String("target", "localhost:3000", "Target address to forward to")
	var expectedApiKey = flag.String("api-key", "default-key", "Expected API key for authentication")
	flag.Parse()

	if *listenPort < 1 || *listenPort > 65535 {
		log.Fatal("Invalid port")
	}

	fmt.Printf("Starting bore server on port %d\n", *listenPort)

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

func handleConnection(conn net.Conn, targetAddr string, expectedApiKey string) {
	defer conn.Close()

	// Read API key
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		log.Println("Failed to read API key")
		return
	}
	apiKey := strings.TrimSpace(scanner.Text())
	if apiKey != expectedApiKey {
		log.Println("Invalid API key from", conn.RemoteAddr())
		return
	}

	// Connect to target
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Println("Failed to connect to target:", err)
		return
	}
	defer targetConn.Close()

	fmt.Println("Forwarding connection from", conn.RemoteAddr(), "to", targetAddr)

	// Forward data in both directions
	go io.Copy(targetConn, conn)
	io.Copy(conn, targetConn)
}
