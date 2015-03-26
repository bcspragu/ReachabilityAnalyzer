package main

import (
	"./bench"
	"fmt"
)

func main() {
	b := bench.NewFromFile("ex1.bench")

	for _, i := range b.Inputs {
		fmt.Println("Input", i)
	}

	for _, o := range b.Outputs {
		fmt.Println("Output", o)
	}

	for _, gate := range b.Gates {
		switch gate.Type() {
		case "AND":
			and := gate.(*bench.And)
			fmt.Println(and.Summary())
		case "NOT":
			not := gate.(*bench.Not)
			fmt.Println(not.Summary())
		case "DFF":
			dff := gate.(*bench.Dff)
			fmt.Println(dff.Summary())
		}
	}
}
