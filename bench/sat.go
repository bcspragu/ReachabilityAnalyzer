package bench

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

type Clause struct {
	Terms   []int
	Comment string
}

func (c *Clause) isComment() bool {
	return c.Comment != ""
}

func (c *Clause) hasTerms() bool {
	return len(c.Terms) > 0
}

func (c *Clause) string() string {
	if c.isComment() {
		return "c " + c.Comment
	} else {
		var str string
		for _, t := range c.Terms {
			str += strconv.Itoa(t) + " "
		}
		return str + "0"
	}
}

func (b *Bench) IsSat() bool {
	cmd := exec.Command("picosat")
	cmd.Stdin = strings.NewReader(b.SatString())
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			stat := status.ExitStatus()
			if stat == 10 { // Satisfiable
				fmt.Println(b.parseOutput(out.String()))
				return true
			} else if stat == 20 { // Unsatisfiable
				return false
			}
		}
	}
	// Means we got an exit code of 0, which we weren't expecting
	return false
}

func (b *Bench) parseOutput(out string) string {
	ports := b.runners[0].satMap
	relevantIDs := make(map[int]string)
	inputState := make([]string, b.unroll)
	gateState := make([]string, b.unroll)

	for _, in := range b.runners[0].Inputs {
		relevantIDs[in.ID()] = "Input"
	}

	for _, g := range b.runners[0].FFs {
		relevantIDs[g.Inputs()[0].ID()] = "State"
	}
	portCount := len(ports)
	// First we remove the first line, which deals with satisfiability
	res := strings.Split(out, "\n")[1:]
	for _, line := range res {
		sp := strings.Split(line, " ")[1:]
		for _, s := range sp {
			n, _ := strconv.Atoi(s)
			originalID := ((abs(n) - 1) % portCount) + 1
			gate, ok := ports[originalID]
			unrolling := ((abs(n) - 1) / portCount)
			if ok {
				bit := " on"
				if n > 0 {
					bit = " off"
				}

				kind := relevantIDs[originalID]
				if kind == "Input" {
					inputState[unrolling] += " " + gate + bit
				} else if kind == "State" {
					gateState[unrolling] += " " + gate + bit
				}
			}
		}
	}
	var buf bytes.Buffer
	for i := range inputState {
		buf.WriteString(fmt.Sprint("State", i+1, ":", inputState[i], gateState[i], "\n"))
	}
	return buf.String()
}

func (b *Bench) SatString() string {
	var buf bytes.Buffer
	clauses := b.asSat()
	buf.WriteString(fmt.Sprintf("p cnf %d %d\n", varCount(clauses), expCount(clauses)))
	for _, clause := range clauses {
		buf.WriteString(clause.string() + "\n")
	}
	return buf.String()
}

func varCount(clauses []Clause) int {
	vars := make(map[int]bool)
	for _, clause := range clauses {
		if clause.hasTerms() {
			for _, t := range clause.Terms {
				vars[abs(t)] = true
			}
		}
	}
	return len(vars)
}

func expCount(clauses []Clause) int {
	c := 0
	for _, clause := range clauses {
		if clause.hasTerms() {
			c++
		}
	}
	return c
}

func (b *Bench) asSat() []Clause {
	clauses := b.initClauses()
	gateCount := len(b.runners[0].portMap)
	for i := 0; i < b.unroll; i++ {
		// The offset is the number of gates in each unrolling, times the cycle we're on
		offset := gateCount * i

		// Add gates for each unrolling
		clauses = addClauses(clauses, []Clause{commentClause("Unrolling number ", i+1)})
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
			clauses = addClauses(clauses, []Clause{commentClause("Connections between unrolling number ", i+1, " and unrolling number ", i+2)})
			for _, g := range b.runners[0].FFs {
				conn := make([]Clause, 2)

				in := g.Inputs()[0].ID() + offset
				out := g.Outputs()[0].ID() + offset + gateCount

				conn[0].Terms = []int{in, -out}
				conn[1].Terms = []int{-in, out}
				clauses = addClauses(clauses, conn)
			}
		}
	}

	clauses = addClauses(clauses, b.endClauses(gateCount*(b.unroll-1)))
	return clauses
}

func (b *Bench) initClauses() []Clause {
	nFF := len(b.runners[0].FFs)
	clauses := make([]Clause, nFF+1)
	clauses[0] = commentClause("Initial conditions")
	for i, g := range b.runners[0].FFs {
		clauses[i+1].Terms = []int{-g.Outputs()[0].ID()}
	}
	return clauses
}

func (b *Bench) endClauses(offset int) []Clause {
	nFF := len(b.runners[0].FFs)
	ffs := b.runners[0].FFs
	clauses := make([]Clause, nFF+1)
	clauses[0] = commentClause("Goal conditions")
	for i, bit := range b.Goal {
		literal := ffs[i].Inputs()[0].ID() + offset

		// Flip the bit if we want it off
		if bit == '0' {
			literal = -literal
		}

		clauses[i+1].Terms = []int{literal}
	}
	return clauses
}

func (b *Bench) andClauses(g Gate, offset int) []Clause {
	clauses := make([]Clause, 3)
	in1 := g.Inputs()[0].ID() + offset
	in2 := g.Inputs()[1].ID() + offset
	out := g.Outputs()[0].ID() + offset

	clauses[0].Terms = []int{in1, -out}
	clauses[1].Terms = []int{in2, -out}
	clauses[2].Terms = []int{-in1, -in2, out}

	return clauses
}

func (b *Bench) notClauses(g Gate, offset int) []Clause {
	clauses := make([]Clause, 2)
	in := g.Inputs()[0].ID() + offset
	out := g.Outputs()[0].ID() + offset

	clauses[0].Terms = []int{-in, -out}
	clauses[1].Terms = []int{in, out}
	return clauses
}

func addClauses(a, b []Clause) []Clause {
	return append(a, b...)
}

func commentClause(a ...interface{}) Clause {
	return Clause{Comment: fmt.Sprint(a...)}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
