package main

import (
	"./bench"
	"flag"
	"fmt"
)

func main() {
	input := flag.String("input", "bench/ex1", "bench file to parse")
	flag.Parse()

	b, _ := bench.NewFromFile(*input)
	fmt.Println(b.ReachableStates())
}
