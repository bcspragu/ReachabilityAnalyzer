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

	//b, _ := bench.NewFromFile("bench/ex1")
	//fmt.Println(b.ReachableStates())
}
