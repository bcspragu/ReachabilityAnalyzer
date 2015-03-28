package bench

import (
	"github.com/wkschwartz/pigosat"
)

func (b *Bench) AsSat(in, out string, unroll int) pigosat.Formula {
	clauses := []pigosat.Clause{}
	for _, gate := range b.runners[0].Gates {
		switch gate.Type() {
		case "AND":

		}
	}
	return clauses
}

func (b *Bench) andClause(g Gate) []pigosat.Clause {
	clauses := make([]pigosat.Clause, 3)
	in1 := pigosat.Literal(g.Inputs()[0].ID())
	in2 := pigosat.Literal(g.Inputs()[1].ID())
	out := pigosat.Literal(g.Outputs()[0].ID())

	clauses[0] = pigosat.Clause{in1, -out}
	clauses[1] = pigosat.Clause{in2, -out}
	clauses[2] = pigosat.Clause{-in1, -in2, out}
	return clauses
}

func (b *Bench) notClause(g Gate) []pigosat.Clause {
	clauses := make([]pigosat.Clause, 2)
	in := pigosat.Literal(g.Inputs()[0].ID())
	out := pigosat.Literal(g.Outputs()[0].ID())

	clauses[0] = pigosat.Clause{-in, -out}
	clauses[1] = pigosat.Clause{in, out}
	return clauses
}
