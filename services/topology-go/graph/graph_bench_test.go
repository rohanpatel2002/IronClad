package graph

import "testing"

func BenchmarkComputeBlastRadius(b *testing.B) {
	g := NewDefault()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.ComputeBlastRadius("payment-api")
	}
}

func BenchmarkDetectCycles(b *testing.B) {
	g := NewDefault()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.DetectCycles()
	}
}
