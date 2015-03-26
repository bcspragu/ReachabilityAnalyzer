package main

import (
	"./bench"
	"fmt"
)

func main() {
	b := bench.NewFromFile("ex2.bench")
	fmt.Println(b.IsReachable("111111111111111"))
	fmt.Println(b.PossibleSolution("111111111111111"))
}
