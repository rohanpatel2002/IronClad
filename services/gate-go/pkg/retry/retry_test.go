package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDoWithExponentialBackoff_Success(t *testing.T) {
	ctx := context.Background()
	count := 0
	operation := func() (interface{}, error) {
		count++
		if count < 3 {
			return nil, errors.New("temporary error")
		}
		return "success", nil
	}

	res, err := DoWithExponentialBackoff(ctx, 3, 10*time.Millisecond, 100*time.Millisecond, operation)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if res != "success" {
		t.Errorf("expected 'success', got %v", res)
	}
	if count != 3 {
		t.Errorf("expected 3 attempts, got %d", count)
	}
}

func TestDoWithExponentialBackoff_MaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()
	count := 0
	operation := func() (interface{}, error) {
		count++
		return nil, errors.New("persistent error")
	}

	_, err := DoWithExponentialBackoff(ctx, 2, 10*time.Millisecond, 100*time.Millisecond, operation)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if count != 3 { // 1 attempt + 2 retries
		t.Errorf("expected 3 attempts, got %d", count)
	}
}

func TestDoWithExponentialBackoff_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	operation := func() (interface{}, error) {
		return nil, errors.New("error")
	}

	// Cancel context after a short delay
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, err := DoWithExponentialBackoff(ctx, 5, 50*time.Millisecond, 200*time.Millisecond, operation)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}
