package main

import (
	"testing"
)

func TestPortValidation(t *testing.T) {
	testCases := []struct {
		name     string
		port     int
		expected bool
	}{
		{"valid port 3000", 3000, true},
		{"valid port 8080", 8080, true},
		{"valid port 1", 1, true},
		{"valid port 65535", 65535, true},
		{"invalid port 0", 0, false},
		{"invalid port -1", -1, false},
		{"invalid port 70000", 70000, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := tc.port >= 1 && tc.port <= 65535
			if isValid != tc.expected {
				t.Errorf("Port %d: expected valid=%v, got valid=%v", tc.port, tc.expected, isValid)
			}
		})
	}
}

func TestServerAddressValidation(t *testing.T) {
	testCases := []struct {
		name     string
		address  string
		expected bool
	}{
		{"valid localhost", "localhost:8080", true},
		{"valid IP", "127.0.0.1:8080", true},
		{"valid hostname", "example.com:8080", true},
		{"invalid no port", "localhost", false},
		{"invalid empty", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Basic validation - just check if it contains a colon
			isValid := len(tc.address) > 0 && contains(tc.address, ":")
			if isValid != tc.expected {
				t.Errorf("Address %s: expected valid=%v, got valid=%v", tc.address, tc.expected, isValid)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
