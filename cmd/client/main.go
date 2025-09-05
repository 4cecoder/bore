package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
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

func handleLocalConnection(localConn net.Conn, serverAddr string, apiKey string) {
	defer localConn.Close()

	// Connect to server for this connection
	serverConn, err := tls.Dial("tcp", serverAddr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		log.Println("Failed to connect to server:", err)
		return
	}
	defer serverConn.Close()

	// Send API key for authentication
	if _, err := fmt.Fprintf(serverConn, "%s\n", apiKey); err != nil {
		log.Println("Failed to send API key:", err)
		return
	}

	fmt.Println("Tunneling connection from", localConn.RemoteAddr())

	// Forward data in both directions
	go io.Copy(serverConn, localConn)
	io.Copy(localConn, serverConn)
}
