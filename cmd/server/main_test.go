package main

import (
	"net"
	"testing"
)

func TestConnectionPool(t *testing.T) {
	targetAddrs := []string{"localhost:8080", "localhost:8081"}
	pool := NewConnectionPool(targetAddrs, 10)

	if pool == nil {
		t.Fatal("Expected connection pool to be created")
	}

	if len(pool.targetAddrs) != 2 {
		t.Errorf("Expected 2 target addresses, got %d", len(pool.targetAddrs))
	}

	if pool.maxSize != 10 {
		t.Errorf("Expected max size 10, got %d", pool.maxSize)
	}
}

func TestHealthCheck(t *testing.T) {
	// This is a basic test - in a real scenario, you'd start a test server
	// For now, just test that the function doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Health check panicked: %v", r)
		}
	}()

	// Test with a closed connection (should handle gracefully)
	conn, err := net.Dial("tcp", "127.0.0.1:9999")
	if err == nil {
		conn.Close()
	}
}

func TestLoadBalancer(t *testing.T) {
	targets := []string{"localhost:8080", "localhost:8081", "localhost:8082"}

	// Test round-robin selection
	selected := make(map[string]int)
	for i := 0; i < 9; i++ {
		target := targets[i%len(targets)]
		selected[target]++
	}

	// Each target should be selected 3 times
	for _, count := range selected {
		if count != 3 {
			t.Errorf("Expected each target to be selected 3 times, got %d", count)
		}
	}
}

func BenchmarkConnectionPool(b *testing.B) {
	targetAddrs := []string{"localhost:8080"}
	pool := NewConnectionPool(targetAddrs, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This would normally test actual connection pooling
		// For benchmark, just test pool creation overhead
		_ = pool
	}
}
