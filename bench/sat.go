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

func (b *Bench) Sat() (bool, string) {
	cmd := exec.Command("picosat")
	cmd.Stdin = strings.NewReader(b.SatString())
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			stat := status.ExitStatus()
			if stat == 10 { // Satisfiable
				return true, b.parseOutput(out.String())
			} else if stat == 20 { // Unsatisfiable
				return false, ""
			}
		}
	}
	// Means we got an exit code of 0, which we weren't expecting
	return false, ""
}

func (b *Bench) parseOutput(out string) string {
	// SatMap maps from ids to gate names, for more verbose output
	relevantIDs := make(map[int]gateType)
	inputState := make([]string, b.Unroll)
	gateState := make([]string, b.Unroll+1) // The plus 1 accounts for initial state

	for _, line := range b.lines {
		if line.gateType == "INPUT" {
			id := b.portMap[line.output]
			gt := relevantIDs[id]
			gt.input = true
			relevantIDs[id] = gt
		} else if line.gateType == "DFF" {
			id := b.portMap[line.inputs[0]]
			gt := relevantIDs[id]
			gt.ff = true
			relevantIDs[id] = gt

			id = b.portMap[line.output]
			gt = relevantIDs[id]
			gt.initial = true
			relevantIDs[id] = gt
		}
	}

	portCount := len(b.portMap)
	// First we remove the first line, which deals with satisfiability
	res := strings.Split(out, "\n")[1:]
	for _, line := range res {
		// Then we remove the leading 'v'
		sp := strings.Split(line, " ")[1:]
		for _, s := range sp {
			n, _ := strconv.Atoi(s)

			originalID := ((abs(n) - 1) % portCount) + 1
			unrolling := ((abs(n) - 1) / portCount)

			if gt, ok := relevantIDs[originalID]; ok {
				bit := "0"
				if n > 0 {
					bit = "1"
				}

				if gt.input {
					inputState[unrolling] += bit
				}

				if gt.ff {
					gateState[unrolling+1] += bit
				}

				// The ID check makes sure that it's the first unrolling
				if gt.initial && abs(n) == originalID {
					gateState[unrolling] += bit
				}
			}
		}
	}
	var buf bytes.Buffer
	for i := range gateState {
		if i == 0 {
			buf.WriteString(fmt.Sprint("Initial: ", gateState[i], " Inputs: ", inputState[i], "\n"))
		} else if i < len(gateState)-1 {
			buf.WriteString(fmt.Sprint("State ", i+1, ": ", gateState[i], " Inputs: ", inputState[i], "\n"))
		} else {
			buf.WriteString(fmt.Sprint("Final: ", gateState[i], "\n"))
		}
	}
	return buf.String()
}

type gateType struct {
	input   bool
	ff      bool
	initial bool
	and     bool
	not     bool
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
	portCount := len(b.portMap)
	for i := 0; i < b.Unroll; i++ {
		// The offset is the number of gates in each unrolling, times the cycle we're on
		offset := portCount * i

		// Add gates for each unrolling
		clauses = addClauses(clauses, []Clause{commentClause("Unrolling number ", i+1)})
		for id := range b.toOutputs {
			// We don't set conditions on ports
			if b.gateType[id].input {
				continue
			} else if b.gateType[id].and {
				clauses = addClauses(clauses, b.andClauses(id, offset))
			} else if b.gateType[id].not {
				clauses = addClauses(clauses, b.notClauses(id, offset))
			}
		}

		// Add connection constraint between unrollings, except the last one
		if i != b.Unroll-1 {
			clauses = addClauses(clauses, []Clause{commentClause("Connections between unrolling number ", i+1, " and unrolling number ", i+2)})
			for _, g := range b.ffs {
				conn := make([]Clause, 2)

				in := b.ports[g].inputs[0] + offset
				out := b.ports[g].output + offset + portCount

				conn[0].Terms = []int{in, -out}
				conn[1].Terms = []int{-in, out}
				clauses = addClauses(clauses, conn)
			}
		}
	}

	clauses = addClauses(clauses, b.endClauses(portCount*(b.Unroll-1)))
	return clauses
}

func (b *Bench) initClauses() []Clause {
	nFF := len(b.ffs)
	clauses := make([]Clause, nFF+1)
	clauses[0] = commentClause("Initial conditions")
	for i, g := range b.ffs {
		clauses[i+1].Terms = []int{-b.ports[g].output}
	}
	return clauses
}

func (b *Bench) endClauses(offset int) []Clause {
	nFF := len(b.ffs)
	clauses := make([]Clause, nFF+1)
	clauses[0] = commentClause("Goal conditions")
	for i, bit := range b.Goal {
		literal := b.ports[b.ffs[i]].inputs[0] + offset

		// Flip the bit if we want it off
		if bit == '0' {
			literal = -literal
		}

		clauses[i+1].Terms = []int{literal}
	}
	return clauses
}

func (b *Bench) andClauses(id, offset int) []Clause {
	clauses := make([]Clause, 3)
	in1 := b.ports[id].inputs[0] + offset
	in2 := b.ports[id].inputs[1] + offset
	out := b.ports[id].output + offset

	clauses[0].Terms = []int{in1, -out}
	clauses[1].Terms = []int{in2, -out}
	clauses[2].Terms = []int{-in1, -in2, out}

	return clauses
}

func (b *Bench) notClauses(id, offset int) []Clause {
	clauses := make([]Clause, 2)
	in := b.ports[id].inputs[0] + offset
	out := b.ports[id].output + offset

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
