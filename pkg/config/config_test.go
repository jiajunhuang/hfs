package config

import (
	"testing"
)

func TestConfig(t *testing.T) {
	if GRPCAddr != "127.0.0.1:8899" {
		t.Fatalf("default GRPCAddr should be 127.0.0.1:8899")
	}
}
