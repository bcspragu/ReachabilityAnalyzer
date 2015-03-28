package bench

import (
	p "github.com/wkschwartz/pigosat"
	"os"
)

func (b *Bench) RunSat() (p.Status, p.Solution, error) {
	pigo := b.pigosat()
	stat, sol := pigo.Solve()
	return stat, sol, nil
}

func (b *Bench) SatToFile(file *os.File) error {
	return b.pigosat().Print(file)
}

func (b *Bench) pigosat() *p.Pigosat {
	pigo, _ := p.New(&p.Options{})
	sat := b.asSat()
	pigo.AddClauses(sat)
	return pigo
}

func (b *Bench) asSat() p.Formula {
	clauses := b.initClauses()
	gateCount := len(b.runners[0].Gates)
	for i := 0; i < b.unroll; i++ {
		// The offset is the number of gates in each unrolling, times the cycle we're on
		offset := gateCount * i

		// Add gates for each unrolling
		for _, gate := range b.runners[0].Gates {
			switch gate.Type() {
			case "AND":
				clauses = addClauses(clauses, b.andClauses(gate, offset))
			case "NOT":
				clauses = addClauses(clauses, b.notClauses(gate, offset))
			}
		}
		// Add connection constraint between unrollings, except the last one
		if i != b.unroll-1 {
			for _, g := range b.runners[0].FFs {
				conn := make([]p.Clause, 2)

				in := p.Literal(g.Inputs()[0].ID() + offset)
				out := p.Literal(g.Outputs()[0].ID() + offset + gateCount)

				conn[0] = p.Clause{in, -out}
				conn[1] = p.Clause{-in, out}
				clauses = addClauses(clauses, conn)
			}
		}
	}

	clauses = addClauses(clauses, b.endClauses(gateCount*(b.unroll-1)))
	return clauses
}

//-4 12 0
//4 -12 0

func (b *Bench) initClauses() []p.Clause {
	nFF := len(b.runners[0].FFs)
	clauses := make([]p.Clause, nFF)
	for i, g := range b.runners[0].FFs {
		clauses[i] = p.Clause{p.Literal(-g.Outputs()[0].ID())}
	}
	return clauses
}

func (b *Bench) endClauses(offset int) []p.Clause {
	nFF := len(b.runners[0].FFs)
	ffs := b.runners[0].FFs
	clauses := make([]p.Clause, nFF)
	for i, bit := range b.Goal {
		literal := p.Literal(ffs[i].Inputs()[0].ID() + offset)

		// Flip the bit if we want it off
		if bit == '0' {
			literal = -literal
		}

		clauses[i] = p.Clause{literal}
	}
	return clauses
}

func (b *Bench) andClauses(g Gate, offset int) []p.Clause {
	clauses := make([]p.Clause, 3)
	in1 := p.Literal(g.Inputs()[0].ID() + offset)
	in2 := p.Literal(g.Inputs()[1].ID() + offset)
	out := p.Literal(g.Outputs()[0].ID() + offset)

	clauses[0] = p.Clause{in1, -out}
	clauses[1] = p.Clause{in2, -out}
	clauses[2] = p.Clause{-in1, -in2, out}
	return clauses
}

func (b *Bench) notClauses(g Gate, offset int) []p.Clause {
	clauses := make([]p.Clause, 2)
	in := p.Literal(g.Inputs()[0].ID() + offset)
	out := p.Literal(g.Outputs()[0].ID() + offset)

	clauses[0] = p.Clause{-in, -out}
	clauses[1] = p.Clause{in, out}
	return clauses
}

func addClauses(a, b []p.Clause) []p.Clause {
	return append(a, b...)
}
