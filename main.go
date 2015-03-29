package main

import (
	"./bench"
	"flag"
	"fmt"
	//"io/ioutil"
	//"os"
)

func main() {
	input := flag.String("input", "bench/ex1", "bench file to parse")
	flag.Parse()

	b, _ := bench.NewFromFile(*input)

	fmt.Println("Symbolic:", b.IsSat())
	fmt.Println("Reachable:", b.IsReachable())
}
