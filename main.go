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

	sat, str := b.Sat()
	fmt.Println("Symbolic:", sat)
	fmt.Print(str)
	//fmt.Println("Reachable:", b.IsReachable())
}
