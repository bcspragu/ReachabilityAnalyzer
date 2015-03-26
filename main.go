package main

import (
	"./bench"
	"fmt"
)

func main() {
	b := bench.NewFromFile("ex1.bench")
	fmt.Println(b.IsReachable("10"))
}
