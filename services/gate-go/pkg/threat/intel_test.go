package threat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIntelClient_IsMalicious(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("1.2.3.4\n5.6.7.8\n10.0.0.0/24\n# comment\n\n9.10.11.12 5\ninvalid-ip-entry"))
	}))
	defer ts.Close()

	client := &IntelClient{
		maliciousIPs: make(map[string]bool),
		client:       http.DefaultClient,
		stop:         make(chan struct{}),
	}

	mockSource1 := &testSource{name: "src1", url: ts.URL, format: "plain"}
	mockSource2 := &testSource{name: "src2", url: ts.URL, format: "ipsum"}
	client.sources = []IntelSource{mockSource1, mockSource2}

	client.refreshFeeds()

	tests := []struct {
		ip       string
		expected bool
	}{
		{"1.2.3.4", true},
		{"9.10.11.12", true},
		{"10.0.0.50", true},
		{"10.0.1.50", false},
		{"0.0.0.0", false},
		{"  1.2.3.4  ", true},
		{"invalid.ip", false},
		{"", false},
	}

	for _, tt := range tests {
		got := client.IsMalicious(tt.ip)
		if got != tt.expected {
			t.Errorf("IsMalicious(%q) = %v, expected %v", tt.ip, got, tt.expected)
		}
	}

	metrics := client.Metrics()
	if metrics.SuccessfulFetchCount != 2 {
		t.Errorf("Expected SuccessfulFetchCount to be 2, got %d", metrics.SuccessfulFetchCount)
	}
	if metrics.TotalIPsFetched != 3 {
		t.Errorf("Expected TotalIPsFetched to be 3, got %d", metrics.TotalIPsFetched)
	}
	if metrics.TotalSubnetsFetched != 1 {
		t.Errorf("Expected TotalSubnetsFetched to be 1, got %d", metrics.TotalSubnetsFetched)
	}
}

func TestIntelClient_TrustedCIDRWhitelist(t *testing.T) {
	client := &IntelClient{
		maliciousIPs: map[string]bool{"1.2.3.4": true},
	}
	if !client.IsMalicious("1.2.3.4") {
		t.Errorf("Expected 1.2.3.4 to be malicious initially")
	}

	if err := client.AddTrustedCIDR("1.2.3.0/24"); err != nil {
		t.Fatalf("Failed to add trusted CIDR: %v", err)
	}

	if client.IsMalicious("1.2.3.4") {
		t.Errorf("Expected 1.2.3.4 to be whitelisted after adding trusted CIDR")
	}
}


func TestIntelSource_FetchWithRetry_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := fetchWithRetry(ctx, http.DefaultClient, ts.URL)
	if err == nil {
		t.Errorf("Expected fetchWithRetry to fail on HTTP 500 error")
	}
}

type testSource struct {
	name   string
	url    string
	format string
}

func (s *testSource) Name() string { return s.name }
func (s *testSource) FetchIPs(ctx context.Context, client *http.Client) (map[string]bool, error) {
	ips := make(map[string]bool)
	if s.format == "plain" {
		ips["1.2.3.4"] = true
		ips["5.6.7.8"] = true
		ips["10.0.0.0/24"] = true
	} else {
		ips["9.10.11.12"] = true
	}
	return ips, nil
}
