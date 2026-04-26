package clients_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sony/gobreaker"
)

// circuitBreakerForTest creates an isolated circuit breaker for unit testing.
func circuitBreakerForTest(name string) *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        name,
		MaxRequests: 1,
		Interval:    time.Second,
		Timeout:     500 * time.Millisecond,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	})
}

func TestCircuitBreaker_AllowsRequestsWhenClosed(t *testing.T) {
	cb := circuitBreakerForTest("test-closed")
	callCount := 0

	for i := 0; i < 5; i++ {
		_, err := cb.Execute(func() (interface{}, error) {
			callCount++
			return "ok", nil
		})
		if err != nil {
			t.Fatalf("expected no error on call %d, got %v", i, err)
		}
	}
	if callCount != 5 {
		t.Fatalf("expected 5 calls, got %d", callCount)
	}
}

func TestCircuitBreaker_OpensAfterConsecutiveFailures(t *testing.T) {
	cb := circuitBreakerForTest("test-open")

	// Trigger 3 consecutive failures
	for i := 0; i < 3; i++ {
		_, _ = cb.Execute(func() (interface{}, error) {
			return nil, http.ErrNoCookie // arbitrary error
		})
	}

	// Now the circuit should be open
	_, err := cb.Execute(func() (interface{}, error) {
		return "should not reach", nil
	})

	if err == nil {
		t.Fatal("expected circuit breaker to be open, but request was allowed through")
	}
}

func TestCircuitBreaker_ClosesAfterTimeout(t *testing.T) {
	cb := circuitBreakerForTest("test-halfopen")

	// Force open
	for i := 0; i < 3; i++ {
		_, _ = cb.Execute(func() (interface{}, error) {
			return nil, http.ErrNoCookie
		})
	}

	// Wait for the timeout to elapse (half-open window)
	time.Sleep(600 * time.Millisecond)

	// MaxRequests=1 so one probe is allowed through in half-open state
	_, err := cb.Execute(func() (interface{}, error) {
		return "recovered", nil
	})
	if err != nil {
		t.Fatalf("expected circuit breaker to allow probe after timeout, got %v", err)
	}
}

func TestHTTPClientWithCircuitBreaker_FailsOnServerError(t *testing.T) {
	// Spin up a test server that always returns 500
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	cb := circuitBreakerForTest("test-http-500")
	client := &http.Client{Timeout: time.Second}

	_, err := cb.Execute(func() (interface{}, error) {
		resp, reqErr := client.Get(ts.URL)
		if reqErr != nil {
			return nil, reqErr
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 500 {
			return nil, context.DeadlineExceeded // treat 5xx as error for CB
		}
		return resp, nil
	})

	if err == nil {
		t.Fatal("expected error on 500 response")
	}
}
