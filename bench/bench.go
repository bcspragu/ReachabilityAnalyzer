package bench

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
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

type Bench struct {
	portMap map[string]*Port

	lastPortID int
	lastGateID int

	Inputs  []*Port
	Outputs []*Port

	Gates []Gate
	FFs   []Gate

	States []string
}

func NewFromFile(filename string) *Bench {
	fmt.Println("Loading bench to memory")

	bench := new(Bench)

	bench.Gates = []Gate{}
	bench.FFs = []Gate{}
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
		line = strings.TrimSpace(line)
		bench.ParseLine(line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
	return bench
}

func (b *Bench) ParseLine(line string) {
	fmt.Println("Parsing line", line)
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
		// Otherwise, it's an input or output
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
	fmt.Println("Adding AND", in1, in2, out)
	inPort1 := b.FindOrCreatePort(in1)
	inPort2 := b.FindOrCreatePort(in2)
	outPort := b.FindOrCreatePort(out)

	and := NewAND(b.nextGateID(), inPort1, inPort2, outPort)
	b.Gates = append(b.Gates, and)
}

func (b *Bench) AddNOT(in, out string) {
	b.addOneInputGate(in, out, NewNOT)
}

func (b *Bench) AddDFF(in, out string) {
	gate := b.addOneInputGate(in, out, NewDFF)
	b.FFs = append(b.FFs, gate)
}

func (b *Bench) AddInput(p string) {
	b.Inputs = append(b.Inputs, b.FindOrCreatePort(p))
}

func (b *Bench) AddOutput(p string) {
	b.Outputs = append(b.Outputs, b.FindOrCreatePort(p))
}

func (b *Bench) addOneInputGate(in, out string, newGate func(int, *Port, *Port) Gate) Gate {
	inPort := b.FindOrCreatePort(in)
	outPort := b.FindOrCreatePort(out)

	gate := newGate(b.nextGateID(), inPort, outPort)
	fmt.Println("Adding", gate.Type(), in, out)
	b.Gates = append(b.Gates, gate)
	return gate
}

func (b *Bench) FindOrCreatePort(name string) *Port {
	if port, ok := b.portMap[name]; ok {
		// Port exists
		return port
	}
	// Create the port if it doesn't exist, using the next
	// available sequential id
	port := NewPort(b.nextPortID())
	b.portMap[name] = port
	return port
}

func (b *Bench) nextPortID() int {
	b.lastPortID++
	return b.lastPortID
}

func (b *Bench) nextGateID() int {
	b.lastGateID++
	return b.lastGateID
}

func (b *Bench) State() string {
	// One character for each flip flop
	buf := make([]byte, 0, len(b.FFs))
	buffer := bytes.NewBuffer(buf)

	for _, dff := range b.FFs {
		if dff.Outputs()[0].on {
			buffer.WriteString("1")
		} else {
			buffer.WriteString("0")
		}
	}

	return buffer.String()
}

func (b *Bench) Summary() string {
	var buffer bytes.Buffer

	for _, i := range b.Inputs {
		buffer.WriteString(fmt.Sprintln("Input", i.ID()))
		for _, conn := range i.Conns() {
			// Go through the inputs of the connection and find
			// which one is connected to the input
			for _, in := range conn.Inputs() {
				if in == i {
					buffer.WriteString(fmt.Sprintln("\tInput to", conn.Type(), "gate ID", conn.ID()))
				}
			}
		}
	}

	for _, o := range b.Outputs {
		buffer.WriteString(fmt.Sprintln("Output", o.ID()))
		for _, conn := range o.Conns() {
			// Go through the inputs of the connection and find
			// which one is connected to the input
			for _, out := range conn.Outputs() {
				if out == o {
					buffer.WriteString(fmt.Sprintln("\tOutput to", conn.Type(), "gate ID", conn.ID()))
				}
			}
		}
	}

	for _, gate := range b.Gates {
		switch gate.Type() {
		case "AND":
			and := gate.(*And)
			buffer.WriteString(fmt.Sprintln(and.Summary()))
		case "NOT":
			not := gate.(*Not)
			buffer.WriteString(fmt.Sprintln(not.Summary()))
		case "DFF":
			dff := gate.(*Dff)
			buffer.WriteString(fmt.Sprintln(dff.Summary()))
		}
	}

	return buffer.String()
}

func (b *Bench) IsReachable(state string) bool {

}

func (b *Bench) Step() {

}
