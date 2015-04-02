package bench

import (
	"runtime"
	"testing"
)

func BenchmarkNewEx1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewFromFile("ex1", 10)
	}
}

func BenchmarkNewEx2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewFromFile("ex2", 10)
	}
}

func BenchmarkNewEx3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewFromFile("ex3", 10)
	}
}

func BenchmarkNewEx4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewFromFile("ex4", 10)
	}
}

func BenchmarkIsReachableEx1(b *testing.B) {
	prev := runtime.GOMAXPROCS(runtime.NumCPU())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isReachable(b, "ex1", 250)
	}
	b.StopTimer()
	runtime.GOMAXPROCS(prev)
}

func BenchmarkIsReachableEx2(b *testing.B) {
	prev := runtime.GOMAXPROCS(runtime.NumCPU())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isReachable(b, "ex2", 250)
	}
	b.StopTimer()
	runtime.GOMAXPROCS(prev)
}

func BenchmarkIsReachableEx3(b *testing.B) {
	prev := runtime.GOMAXPROCS(runtime.NumCPU())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isReachable(b, "ex3", 250)
	}
	b.StopTimer()
	runtime.GOMAXPROCS(prev)
}

func BenchmarkIsReachableEx4(b *testing.B) {
	prev := runtime.GOMAXPROCS(runtime.NumCPU())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isReachable(b, "ex4", 250)
	}
	b.StopTimer()
	runtime.GOMAXPROCS(prev)
}

func BenchmarkRunLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		nextState(b, "ex3", "00000000000000000000000000000", "1111111111000111")
	}
}

func BenchmarkEx1Sat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex1", 10)
	}
}

func BenchmarkEx2Sat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex2", 10)
	}
}

func BenchmarkEx3Sat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex3", 10)
	}
}

func BenchmarkEx4Sat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 10)
	}
}

func BenchmarkEx4Unroll1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 1)
	}
}

func BenchmarkEx4Unroll2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 2)
	}
}

func BenchmarkEx4Unroll3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 3)
	}
}

func BenchmarkEx4Unroll4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 4)
	}
}

func BenchmarkEx4Unroll5(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 5)
	}
}

func BenchmarkEx4Unroll6(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 6)
	}
}

func BenchmarkEx4Unroll7(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 7)
	}
}

func BenchmarkEx4Unroll8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 8)
	}
}

func BenchmarkEx4Unroll9(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 9)
	}
}

func BenchmarkEx4Unroll10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 10)
	}
}

func BenchmarkEx4Unroll11(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 11)
	}
}

func BenchmarkEx4Unroll12(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 12)
	}
}

func BenchmarkEx4Unroll13(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 13)
	}
}

func BenchmarkEx4Unroll14(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 14)
	}
}

func BenchmarkEx4Unroll15(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 15)
	}
}

func BenchmarkEx4Unroll16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 16)
	}
}

func BenchmarkEx4Unroll17(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 17)
	}
}

func BenchmarkEx4Unroll18(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 18)
	}
}

func BenchmarkEx4Unroll19(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 19)
	}
}

func BenchmarkEx4Unroll20(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runSat(b, "ex4", 20)
	}
}

func TestRun(t *testing.T) {
	init := "00000000000000000000000000000"
	input := "1111111111000111"
	exp := "00110100010000010011000000100"

	bench, _ := NewFromFile("ex3", 1)
	if s := bench.NextState(init, input); s != exp {
		t.Errorf("Expected %s, Got %s", exp, s)
	}
}

func nextState(b *testing.B, in, state, input string) {
	b.StopTimer()
	bench, _ := NewFromFile(in, 1)
	b.StartTimer()
	bench.NextState(state, input)
}

func runSat(b *testing.B, in string, unroll int) {
	b.StopTimer()
	bench, _ := NewFromFile(in, 0)
	bench.Unroll = unroll
	b.StartTimer()
	bench.Sat()
}

func isReachable(b *testing.B, in string, runners int) {
	b.StopTimer()
	bench, _ := NewFromFile(in, runners)
	b.StartTimer()
	bench.IsReachable()
}
