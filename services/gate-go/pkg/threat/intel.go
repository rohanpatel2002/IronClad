package threat

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	abuseCHURL = "https://feodotracker.abuse.ch/downloads/ipblocklist.txt"
)

// IntelSource defines the interface for different threat intelligence providers.
type IntelSource interface {
	Name() string
	FetchIPs(ctx context.Context, client *http.Client) (map[string]bool, error)
}

// AbuseCHSource implements IntelSource for abuse.ch
type AbuseCHSource struct{}

func (s *AbuseCHSource) Name() string { return "abuse.ch" }
func (s *AbuseCHSource) FetchIPs(ctx context.Context, client *http.Client) (map[string]bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, abuseCHURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "IronClad-ThreatIntel-Bot/1.0 (https://github.com/rohanpatel2002/IronClad)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	ips := make(map[string]bool)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ips[line] = true
	}
	return ips, scanner.Err()
}

// IpsumSource implements IntelSource for ipsum (aggregated list)
type IpsumSource struct{}

func (s *IpsumSource) Name() string { return "ipsum" }
func (s *IpsumSource) FetchIPs(ctx context.Context, client *http.Client) (map[string]bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://raw.githubusercontent.com/stamparm/ipsum/master/ipsum.txt", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "IronClad-ThreatIntel-Bot/1.0 (https://github.com/rohanpatel2002/IronClad)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	ips := make(map[string]bool)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// ipsum format is "IP LEVEL", we just want IP
		parts := strings.Fields(line)
		if len(parts) > 0 {
			ips[parts[0]] = true
		}
	}
	return ips, scanner.Err()
}

// IntelClient pulls global threat feeds from multiple sources.
type IntelClient struct {
	mu           sync.RWMutex
	maliciousIPs map[string]bool
	maliciousSubnets []*net.IPNet
	client       *http.Client
	sources      []IntelSource
	stop         chan struct{}
}

// NewIntelClient initializes a client with multiple sources.
func NewIntelClient() *IntelClient {
	c := &IntelClient{
		maliciousIPs: make(map[string]bool),
		client:       &http.Client{Timeout: 10 * time.Second},
		sources:      []IntelSource{&AbuseCHSource{}, &IpsumSource{}},
		stop:         make(chan struct{}),
	}
	// Initial fetch
	c.refreshFeeds()
	go c.refreshLoop()
	return c
}

// Stop shuts down the background refresh loop.
func (c *IntelClient) Stop() {
	close(c.stop)
}

func (c *IntelClient) refreshLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.refreshFeeds()
		case <-c.stop:
			slog.Info("Threat Intel refresh loop stopped")
			return
		}
	}
}

func (c *IntelClient) refreshFeeds() {
	newIPs := make(map[string]bool)
	var newSubnets []*net.IPNet
	var subnetMu sync.Mutex
	var wg sync.WaitGroup
	var mu sync.Mutex

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, s := range c.sources {
		wg.Add(1)
		go func(src IntelSource) {
			defer wg.Done()
			ips, err := src.FetchIPs(ctx, c.client)
			if err != nil {
				slog.Error("Error fetching from threat source", "source", src.Name(), "error", err)
				return
			}
			mu.Lock()
			for ip := range ips {
				if strings.Contains(ip, "/") {
					if _, ipnet, err := net.ParseCIDR(ip); err == nil {
						subnetMu.Lock()
						newSubnets = append(newSubnets, ipnet)
						subnetMu.Unlock()
					}
				} else {
					newIPs[ip] = true
				}
			}
			mu.Unlock()
			slog.Debug("Fetched IPs from threat source", "count", len(ips), "source", src.Name())
		}(s)
	}

	wg.Wait()

	c.mu.Lock()
	c.maliciousIPs = newIPs
	c.maliciousSubnets = newSubnets
	c.mu.Unlock()
	slog.Info("Threat database updated", "total_malicious_ips", len(newIPs), "total_malicious_subnets", len(newSubnets))
}

// IsMalicious returns true if the IP is found in the global threat database or falls within a blacklisted subnet.
func (c *IntelClient) IsMalicious(ip string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.maliciousIPs[ip] {
		return true
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, subnet := range c.maliciousSubnets {
		if subnet.Contains(parsedIP) {
			return true
		}
	}

	return false
}
