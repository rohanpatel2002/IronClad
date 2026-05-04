package threat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// IntelClient pulls global threat feeds to identify malicious actors.
type IntelClient struct {
	mu           sync.RWMutex
	maliciousIPs map[string]bool
	client       *http.Client
}

// NewIntelClient initializes a client with a background refresh loop.
func NewIntelClient() *IntelClient {
	c := &IntelClient{
		maliciousIPs: make(map[string]bool),
		client:       &http.Client{Timeout: 5 * time.Second},
	}
	go c.refreshLoop()
	return c
}

func (c *IntelClient) refreshLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		c.refreshFeeds()
	}
}

func (c *IntelClient) refreshFeeds() {
	// Mocking a call to a threat intel feed (e.g., AbuseIPDB, AlienVault)
	newIPs := map[string]bool{
		"1.2.3.4":   true,
		"8.8.8.8":   false, // Example of clean IP
		"192.168.1.1": false,
	}

	c.mu.Lock()
	c.maliciousIPs = newIPs
	c.mu.Unlock()
}

// IsMalicious returns true if the IP is found in the global threat database.
func (c *IntelClient) IsMalicious(ip string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maliciousIPs[ip]
}
