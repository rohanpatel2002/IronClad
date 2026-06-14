package threat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIntelClient_IsMalicious(t *testing.T) {
	// Mock server to return IPs and CIDRs
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("1.2.3.4\n5.6.7.8\n10.0.0.0/24\n# comment\n\n9.10.11.12 5"))
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

	if !client.IsMalicious("1.2.3.4") {
		t.Errorf("Expected 1.2.3.4 to be malicious")
	}
	if !client.IsMalicious("9.10.11.12") {
		t.Errorf("Expected 9.10.11.12 to be malicious")
	}
	if !client.IsMalicious("10.0.0.50") {
		t.Errorf("Expected 10.0.0.50 (in 10.0.0.0/24) to be malicious")
	}
	if client.IsMalicious("10.0.1.50") {
		t.Errorf("Expected 10.0.1.50 (outside 10.0.0.0/24) NOT to be malicious")
	}
	if client.IsMalicious("0.0.0.0") {
		t.Errorf("Expected 0.0.0.0 NOT to be malicious")
	}
}

type testSource struct {
	name   string
	url    string
	format string
}

func (s *testSource) Name() string { return s.name }
func (s *testSource) FetchIPs(ctx context.Context, client *http.Client) (map[string]bool, error) {
	resp, err := client.Get(s.url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	ips := make(map[string]bool)
	if s.format == "plain" {
		ips["1.2.3.4"] = true
		ips["10.0.0.0/24"] = true
	} else {
		ips["9.10.11.12"] = true
	}
	return ips, nil
}

