package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// Config holds the paths to the mTLS certificates.
type Config struct {
	CACertFile string
	CertFile   string
	KeyFile    string
	ServerName string
}

// LoadTLSConfig loads the certificates and returns a *tls.Config.
func LoadTLSConfig(cfg Config) (*tls.Config, error) {
	if cfg.CertFile == "" || cfg.KeyFile == "" {
		return nil, nil // mTLS disabled
	}

	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert/key: %w", err)
	}

	caCert, err := os.ReadFile(cfg.CACertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("failed to append CA certs")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   cfg.ServerName,
	}, nil
}
