package main

import (
	"./bench"
	"flag"
	"fmt"
	"runtime"
)

var (
	nRunners int
	logLevel int
	nUnroll  int

	inputFile string

	explicit bool
	symbolic bool
)

func init() {
	// Flag parsing things
	flag.IntVar(&logLevel, "log", bench.None, "level of verboseness in output")
	flag.IntVar(&nRunners, "runners", 10, "how many processes to simultaneously calculate state")
	flag.IntVar(&nUnroll, "unroll", 2, "how many times to unroll the formula")

	flag.StringVar(&inputFile, "input", "bench/ex1", "bench file to parse")

	flag.BoolVar(&explicit, "e", false, "run explicit search on the input file")
	flag.BoolVar(&symbolic, "s", false, "run symbolic search on the input file")

	flag.Parse()

	// Run on all available cores
	runtime.GOMAXPROCS(runtime.NumCPU())
}
func main() {
	b, _ := bench.NewFromFile(inputFile, nRunners)
	b.LogLevel = logLevel
	b.Unroll = nUnroll
	if explicit {
		reachable, count := b.IsReachable()
		fmt.Println("Explicitly Reachable:", reachable)
		fmt.Println("Number of states found:", count)
	}

	if symbolic {
		sat, _ := b.Sat()
		fmt.Println("Symbolicly Reachable in", nUnroll, "unrollings:", sat)
	}
}
