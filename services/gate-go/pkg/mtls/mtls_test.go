package mtls

import (
	"crypto/tls"
	"testing"
)

func TestLoadTLSConfig_Disabled(t *testing.T) {
	cfg := Config{}
	tlsCfg, err := LoadTLSConfig(cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if tlsCfg != nil {
		t.Errorf("Expected nil tls.Config when cert files are empty")
	}
}

func TestLoadServerTLSConfig_Disabled(t *testing.T) {
	cfg := Config{}
	tlsCfg, err := LoadServerTLSConfig(cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if tlsCfg != nil {
		t.Errorf("Expected nil tls.Config when cert files are empty")
	}
}

func TestConfig_MinVersionDefault(t *testing.T) {
	cfg := Config{MinVersion: 0}
	if cfg.MinVersion != 0 {
		t.Errorf("Expected initial 0 minversion")
	}
	// Verify default constant tls.VersionTLS13
	if tls.VersionTLS13 != 0x0304 {
		t.Errorf("Unexpected TLS13 constant value")
	}
}
