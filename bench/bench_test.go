package bench

import (
	"runtime"
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

func BenchmarkIsReachableEx2Parallel(b *testing.B) {
	prev := runtime.GOMAXPROCS(runtime.NumCPU())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isReachable(b, "ex2")
	}
	b.StopTimer()
	runtime.GOMAXPROCS(prev)
}

func BenchmarkIsReachableEx3Parallel(b *testing.B) {
	prev := runtime.GOMAXPROCS(runtime.NumCPU())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isReachable(b, "ex3")
	}
	b.StopTimer()
	runtime.GOMAXPROCS(prev)
}

func BenchmarkRunLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		nextState(b, "ex3", "00000000000000000000000000000", "1111111111000111")
	}
}

func TestRun(t *testing.T) {
	init := "00000000000000000000000000000"
	input := "1111111111000111"
	exp := "00110100010000010011000000100"

	bench, _ := NewFromFile("ex3")
	if s := bench.NextState(init, input); s != exp {
		t.Errorf("Expected %s, Got %s", exp, s)
	}
}

func nextState(b *testing.B, in, state, input string) {
	b.StopTimer()
	bench, _ := NewFromFile(in)
	b.StartTimer()
	bench.NextState(state, input)
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
