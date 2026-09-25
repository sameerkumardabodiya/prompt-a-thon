package main

import (
	"testing"
)

// TestHealthCheck ensures the coordinator correctly flags unreachable nodes
func TestCheckHealth(t *testing.T) {
	status := checkHealth("http://127.0.0.1:9999") // Intentionally invalid port
	if status != "OFFLINE" {
		t.Errorf("Expected OFFLINE for unreachable node, got %s", status)
	}
}
