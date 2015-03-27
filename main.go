package main

import (
	"./bench"
	"fmt"
	"os"
)

func main() {
	b, _ := bench.NewFromFile(os.Args[1])
	fmt.Println(b.IsReachable())
	sol, err := b.PossibleSolution()
	fmt.Println(sol, err)
}
