package sync

import (
	"testing"
)

func TestDistLock_KeyGeneration(t *testing.T) {
	lockKey := "payment-process"
	lock := NewDistLock(nil, lockKey)

	if lock.key != "lock:payment-process" {
		t.Errorf("Expected lock key 'lock:payment-process', got %s", lock.key)
	}
	if lock.value == "" {
		t.Errorf("Expected non-empty random UUID lock value")
	}
}
