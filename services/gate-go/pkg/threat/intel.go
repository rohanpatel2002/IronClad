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
	"sync/atomic"
	"time"

	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/retry"
)

const (
	abuseCHURL = "https://feodotracker.abuse.ch/downloads/ipblocklist.txt"
	ipsumURL   = "https://raw.githubusercontent.com/stamparm/ipsum/master/ipsum.txt"
)

// IntelMetrics holds statistics and operational counters for the threat intel client.
type IntelMetrics struct {
	TotalIPsFetched      int64
	TotalSubnetsFetched  int64
	FailedFetchCount     int64
	SuccessfulFetchCount int64
	LastRefreshUnix      int64
}

// IntelSource defines the interface for different threat intelligence providers.
type IntelSource interface {
	Name() string
	FetchIPs(ctx context.Context, client *http.Client) (map[string]bool, error)
}

// fetchWithRetry executes HTTP fetch requests using exponential backoff retry logic.
func fetchWithRetry(ctx context.Context, client *http.Client, urlStr string) (*http.Response, error) {
	res, err := retry.DoWithExponentialBackoff(ctx, 3, 200*time.Millisecond, 2*time.Second, func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "IronClad-ThreatIntel-Bot/1.0 (https://github.com/rohanpatel2002/IronClad)")
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
		}
		return resp, nil
	})
	if err != nil {
		return nil, err
	}
	return res.(*http.Response), nil
}

// AbuseCHSource implements IntelSource for abuse.ch
type AbuseCHSource struct{}

func (s *AbuseCHSource) Name() string { return "abuse.ch" }
func (s *AbuseCHSource) FetchIPs(ctx context.Context, client *http.Client) (map[string]bool, error) {
	resp, err := fetchWithRetry(ctx, client, abuseCHURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
	resp, err := fetchWithRetry(ctx, client, ipsumURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ips := make(map[string]bool)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) > 0 {
			ips[parts[0]] = true
		}
	}
	return ips, scanner.Err()
}

// IntelClient pulls global threat feeds from multiple sources.
type IntelClient struct {
	mu               sync.RWMutex
	maliciousIPs     map[string]bool
	maliciousSubnets []*net.IPNet
	trustedSubnets   []*net.IPNet
	client           *http.Client
	sources          []IntelSource
	stop             chan struct{}
	metrics          IntelMetrics
}

// NewIntelClient initializes a client with multiple sources.
func NewIntelClient() *IntelClient {
	c := &IntelClient{
		maliciousIPs:   make(map[string]bool),
		trustedSubnets: make([]*net.IPNet, 0),
		client:         &http.Client{Timeout: 10 * time.Second},
		sources:        []IntelSource{&AbuseCHSource{}, &IpsumSource{}},
		stop:           make(chan struct{}),
	}
	c.refreshFeeds()
	go c.refreshLoop()
	return c
}

// AddTrustedCIDR registers a CIDR subnet that should never be marked as malicious.
func (c *IntelClient) AddTrustedCIDR(cidr string) error {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.trustedSubnets = append(c.trustedSubnets, ipnet)
	return nil
}

// Stop shuts down the background refresh loop.
func (c *IntelClient) Stop() {
	close(c.stop)
}

// Metrics returns a snapshot of current threat intel operational statistics.
func (c *IntelClient) Metrics() IntelMetrics {
	return IntelMetrics{
		TotalIPsFetched:      atomic.LoadInt64(&c.metrics.TotalIPsFetched),
		TotalSubnetsFetched:  atomic.LoadInt64(&c.metrics.TotalSubnetsFetched),
		FailedFetchCount:     atomic.LoadInt64(&c.metrics.FailedFetchCount),
		SuccessfulFetchCount: atomic.LoadInt64(&c.metrics.SuccessfulFetchCount),
		LastRefreshUnix:      atomic.LoadInt64(&c.metrics.LastRefreshUnix),
	}
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
				atomic.AddInt64(&c.metrics.FailedFetchCount, 1)
				slog.Error("Error fetching from threat source", "source", src.Name(), "error", err)
				return
			}
			atomic.AddInt64(&c.metrics.SuccessfulFetchCount, 1)

			mu.Lock()
			for ip := range ips {
				if strings.Contains(ip, "/") {
					if _, ipnet, err := net.ParseCIDR(ip); err == nil {
						subnetMu.Lock()
						newSubnets = append(newSubnets, ipnet)
						subnetMu.Unlock()
					}
				} else {
					if net.ParseIP(ip) != nil {
						newIPs[ip] = true
					}
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

	atomic.StoreInt64(&c.metrics.TotalIPsFetched, int64(len(newIPs)))
	atomic.StoreInt64(&c.metrics.TotalSubnetsFetched, int64(len(newSubnets)))
	atomic.StoreInt64(&c.metrics.LastRefreshUnix, time.Now().Unix())

	slog.Info("Threat database updated", "total_malicious_ips", len(newIPs), "total_malicious_subnets", len(newSubnets))
}

// IsMalicious returns true if the IP is valid and found in the global threat database or falls within a blacklisted subnet.
// It returns false if the IP matches a trusted CIDR whitelist.
func (c *IntelClient) IsMalicious(ip string) bool {
	trimmed := strings.TrimSpace(ip)
	parsedIP := net.ParseIP(trimmed)
	if parsedIP == nil {
		return false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check trusted subnets first (whitelist takes priority)
	for _, trusted := range c.trustedSubnets {
		if trusted.Contains(parsedIP) {
			return false
		}
	}

	if c.maliciousIPs[trimmed] {
		return true
	}

	for _, subnet := range c.maliciousSubnets {
		if subnet.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// ForceRefreshFeeds triggers an immediate synchronous update of all threat feeds.
func (c *IntelClient) ForceRefreshFeeds() {
	c.refreshFeeds()
}


