package main

import (
	"./bench"
	"flag"
	"fmt"
	"os"
)

func main() {
	input := flag.String("input", "bench/ex1", "bench file to parse")
	flag.Parse()

	b, _ := bench.NewFromFile(*input)
	file, _ := os.Create("satty.sat")
	fmt.Println(b.PortMap())
	fmt.Println(b.SatToFile(file))
}
