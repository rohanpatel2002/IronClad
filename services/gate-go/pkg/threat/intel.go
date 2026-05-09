package threat

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	feedURL = "https://feodotracker.abuse.ch/downloads/ipblocklist.txt"
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
		client:       &http.Client{Timeout: 10 * time.Second},
	}
	// Initial fetch
	c.refreshFeeds()
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
	resp, err := c.client.Get(feedURL)
	if err != nil {
		fmt.Printf("Error fetching threat feed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code from threat feed: %d\n", resp.StatusCode)
		return
	}

	newIPs := make(map[string]bool)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		newIPs[line] = true
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading threat feed: %v\n", err)
		return
	}

	c.mu.Lock()
	c.maliciousIPs = newIPs
	c.mu.Unlock()
	fmt.Printf("Refreshed threat intelligence: %d malicious IPs tracked\n", len(newIPs))
}

// IsMalicious returns true if the IP is found in the global threat database.
func (c *IntelClient) IsMalicious(ip string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maliciousIPs[ip]
}
