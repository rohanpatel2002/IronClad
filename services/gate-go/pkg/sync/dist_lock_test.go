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

func TestDistLock_GetKey(t *testing.T) {
	lock := NewDistLock(nil, "order-sync")
	if lock.GetKey() != "lock:order-sync" {
		t.Errorf("expected GetKey() 'lock:order-sync', got %s", lock.GetKey())
	}
}

