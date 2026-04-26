package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/retry"
)

func TestDoWithExponentialBackoff_SucceedsFirstTry(t *testing.T) {
	calls := 0
	res, err := retry.DoWithExponentialBackoff(context.Background(), 3, time.Millisecond, 10*time.Millisecond,
		func() (interface{}, error) {
			calls++
			return "ok", nil
		})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
	if res.(string) != "ok" {
		t.Fatalf("unexpected result: %v", res)
	}
}

func TestDoWithExponentialBackoff_RetriesOnError(t *testing.T) {
	calls := 0
	sentinel := errors.New("transient")

	_, err := retry.DoWithExponentialBackoff(context.Background(), 3, time.Millisecond, 10*time.Millisecond,
		func() (interface{}, error) {
			calls++
			if calls < 3 {
				return nil, sentinel
			}
			return "recovered", nil
		})

	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDoWithExponentialBackoff_ExceedsMaxRetries(t *testing.T) {
	calls := 0
	sentinel := errors.New("always fails")

	_, err := retry.DoWithExponentialBackoff(context.Background(), 3, time.Millisecond, 10*time.Millisecond,
		func() (interface{}, error) {
			calls++
			return nil, sentinel
		})

	if err == nil {
		t.Fatal("expected error after max retries, got nil")
	}
	if calls != 4 { // initial + 3 retries
		t.Fatalf("expected 4 calls (1 + 3 retries), got %d", calls)
	}
}

func TestDoWithExponentialBackoff_RespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0

	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	_, err := retry.DoWithExponentialBackoff(ctx, 10, 10*time.Millisecond, 100*time.Millisecond,
		func() (interface{}, error) {
			calls++
			return nil, errors.New("fail")
		})

	if err == nil {
		t.Fatal("expected context cancellation error")
	}
	// Should not have run all 11 attempts
	if calls >= 11 {
		t.Fatalf("expected fewer than 11 calls due to context cancellation, got %d", calls)
	}
}
