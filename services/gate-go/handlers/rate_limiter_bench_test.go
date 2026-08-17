package handlers

import (
	"testing"
	"time"
)

func BenchmarkLeakyBucket(b *testing.B) {
	lb := NewLeakyBucket(10*time.Millisecond, 100)
	defer lb.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = lb.GetBucketState("127.0.0.1")
		}
	})
}
