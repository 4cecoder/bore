package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	var listenPort = flag.Int("port", 8080, "Port to listen on")
	var targetAddr = flag.String("target", "localhost:3000", "Target address to forward to")
	flag.Parse()

	if *listenPort < 1 || *listenPort > 65535 {
		log.Fatal("Invalid port")
	}

	fmt.Printf("Starting bore server on port %d\n", *listenPort)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *listenPort))
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
		go handleConnection(conn, *targetAddr)
	}
}

func handleConnection(conn net.Conn, targetAddr string) {
	defer conn.Close()

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
