package bench

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

// This beautifully crafted regular expression will match
// statements of the form:
// V0 = AND(A,X1)
// V1 = NOT(X1)
// V2 = DFF(V1)
// V3 = DFF(V0)
// The regex will capture the output port name, the gate type,
// and the input port name(s). The examples above would yield
// the following captures:
// V0, AND, A, X1
// V1, NOT, X1
// V2, DFF, V1
// V3, DFF, V0
// And this is all we need to build up our network
var gateRE = regexp.MustCompile(`^(\w+)\s*=\s*(\w+)\((\w+)(?:,(\w+))*\)$`)
var inOutRE = regexp.MustCompile(`^(\w+)\((\w+)\)$`)

type Gate interface {
	SetOut()
	Out() bool
	Type() string
}

type Bench struct {
	portMap map[string]*Port

	lastID int

	Inputs  []*Port
	Outputs []*Port

	Gates []Gate
}

func NewFromFile(filename string) *Bench {
	fmt.Println("Loading bench to memory")

	bench := new(Bench)

	bench.Gates = []Gate{}
	bench.portMap = make(map[string]*Port)

	bench.Inputs = []*Port{}
	bench.Outputs = []*Port{}

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println("Parsing line", line)
		bench.LoadLine(line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
	return bench
}

func (b *Bench) LoadLine(line string) {
	matches := gateRE.FindStringSubmatch(line)

	// If we have matches, it's a gate statement
	if len(matches) > 0 {
		out, gate := matches[1], matches[2]
		switch gate {
		case "AND":
			b.AddAND(matches[3], matches[4], out)
		case "NOT":
			b.AddNOT(matches[3], out)
		case "DFF":
			b.AddDFF(matches[3], out)
		}
	} else {
		if inOutRE.MatchString(line) {
			ioMatch := inOutRE.FindStringSubmatch(line)
			io, port := ioMatch[1], ioMatch[2]
			switch io {
			case "INPUT":
				b.AddInput(port)
			case "OUTPUT":
				b.AddOutput(port)
			}
		}
	}
}

func (b *Bench) AddAND(in1, in2, out string) {
	inPort1 := b.FindOrCreatePort(in1)
	inPort2 := b.FindOrCreatePort(in2)
	outPort := b.FindOrCreatePort(out)

	and := NewAND(inPort1, inPort2, outPort)
	b.Gates = append(b.Gates, and)
}

func (b *Bench) AddNOT(in, out string) {
	b.addOneInputGate(in, out, NewNOT)
}

func (b *Bench) AddDFF(in, out string) {
	b.addOneInputGate(in, out, NewDFF)
}

func (b *Bench) AddInput(p string) {
	b.Inputs = append(b.Inputs, b.FindOrCreatePort(p))
}

func (b *Bench) AddOutput(p string) {
	b.Outputs = append(b.Outputs, b.FindOrCreatePort(p))
}

func (b *Bench) addOneInputGate(in, out string, newFunc func(*Port, *Port) Gate) {
	inPort := b.FindOrCreatePort(in)
	outPort := b.FindOrCreatePort(out)

	gate := newFunc(inPort, outPort)
	b.Gates = append(b.Gates, gate)
}

func (b *Bench) FindOrCreatePort(name string) *Port {
	if port, ok := b.portMap[name]; ok {
		// Port exists
		return port
	}
	// Create the port if it doesn't exist, using the next
	// available sequential id
	port := NewPort(b.nextID())
	b.portMap[name] = port
	return port
}

func (b *Bench) nextID() int {
	b.lastID++
	return b.lastID
}
