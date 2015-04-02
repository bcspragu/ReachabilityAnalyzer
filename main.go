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
	count    bool
)

func init() {
	// Flag parsing things
	flag.IntVar(&logLevel, "log", bench.None, "level of verboseness in output")
	flag.IntVar(&nRunners, "runners", 10, "how many processes to simultaneously calculate state")
	flag.IntVar(&nUnroll, "unroll", 2, "how many times to unroll the formula")

	flag.StringVar(&inputFile, "input", "bench/ex1", "bench file to parse")

	flag.BoolVar(&explicit, "e", false, "run explicit search on the input file")
	flag.BoolVar(&count, "c", false, "explicitly search for all reachable states and return a count")
	flag.BoolVar(&symbolic, "s", false, "run symbolic search on the input file")

	flag.Parse()

	// Run on all available cores
	runtime.GOMAXPROCS(runtime.NumCPU())
}
func main() {
	b, _ := bench.NewFromFile(inputFile, nRunners)
	b.LogLevel = logLevel
	b.Unroll = nUnroll
	if explicit && count {
		reachable := b.ReachableStates()
		var isReachable bool
		for state := range reachable {
			if state == b.Goal {
				isReachable = true
				break
			}
		}
		fmt.Println("Explicitly Reachable:", isReachable)
		fmt.Println("Total reachable states:", len(reachable))
		if isReachable {
			fmt.Println(b.Solution(reachable))
		}
	} else if explicit && !count {
		isReachable, reachable := b.IsReachable()
		fmt.Println("Explicitly reachable:", isReachable)
		fmt.Println("Number of states found before terminating:", len(reachable))
		if isReachable {
			fmt.Println(b.Solution(reachable))
		}
	} else if count && !explicit {
		reachable := b.ReachableStates()
		fmt.Println("Total reachable states:", len(reachable))
	}

	if symbolic {
		sat, sol := b.Sat()
		fmt.Println("Symbolically reachable in", nUnroll, "unrollings:", sat)
		if sat {
			fmt.Println(sol)
		}
	}
}
