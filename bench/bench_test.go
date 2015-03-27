package bench

import (
	"testing"
)

func BenchmarkNewEx1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		newBench(b, "ex1")
	}
}

func BenchmarkNewEx2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		newBench(b, "ex2")
	}
}

func BenchmarkNewEx3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		newBench(b, "ex3")
	}
}

func BenchmarkIsReachableEx1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isReachable(b, "ex1")
	}
}

func BenchmarkIsReachableEx2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isReachable(b, "ex2")
	}
}

func BenchmarkIsReachableEx3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isReachable(b, "ex3")
	}
}

func isReachable(b *testing.B, in string) {
	b.StopTimer()
	bench, _ := NewFromFile(in)
	b.StartTimer()
	bench.IsReachable()
}

func newBench(b *testing.B, in string) {
	NewFromFile(in)
}
