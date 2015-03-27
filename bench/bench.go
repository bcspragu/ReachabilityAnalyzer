package bench

import (
	"bufio"
	"bytes"
	//"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// This beautifully crafted regular expression will match
// statements of the form:
// V0 = AND(A,X1)
// V1 = NOT(X1)
// V2 = DFF(V1) // V3 = DFF(V0)
// The regex will capture the output port name, the gate type,
// and the input port name(s). The examples above would yield
// the following captures:
// V0, AND, A, X1
// V1, NOT, X1
// V2, DFF, V1
// V3, DFF, V0
// And this is all we need to build up our network
const (
	NONE = iota
	DEBUG
	VERBOSE
)

var gateRE = regexp.MustCompile(`^(\w+)\s*=\s*(\w+)\((\w+)\s*(?:,\s*(\w+))?\)$`)
var inOutRE = regexp.MustCompile(`^(\w+)\((\w+)\)$`)
var verbose int

func init() {
	flag.IntVar(&verbose, "v", NONE, "level of verboseness in output")
}

type Bench struct {
	// PortMap is a map from names in a bench file to their
	// corresponding port objects
	portMap map[string]*Port
	// NextStates is a map from a state string to all of the
	// states that can be reached from it, and what inputs are
	// needed to reach them
	nextStates map[string][]string

	lastPortID int
	lastGateID int

	Inputs  []*Port
	Outputs []*Port

	Gates []Gate
	FFs   []Gate

	States []string
	Goal   string
}

func NewFromFile(filename string) (*Bench, error) {
	debugStatement("Creating bench", VERBOSE)
	bench := new(Bench)

	bench.Gates = []Gate{}
	bench.FFs = []Gate{}
	bench.portMap = make(map[string]*Port)
	bench.nextStates = make(map[string][]string)

	bench.Inputs = []*Port{}
	bench.Outputs = []*Port{}
	goalState, err := ioutil.ReadFile(filename + ".state")
	if err != nil {
		return bench, err
	}
	bench.Goal = strings.TrimSpace(string(goalState))

	file, err := os.Open(filename + ".bench")
	if err != nil {
		return bench, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		bench.ParseLine(line)
	}

	if err := scanner.Err(); err != nil {
		return bench, err
	}
	debugStatement(fmt.Sprint("Bench created with ", len(bench.Gates), " gates"), DEBUG)
	return bench, nil
}

func (b *Bench) ParseLine(line string) {
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
		if dff.Inputs()[0].on {
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

func (b *Bench) ReachableStates() []string {
	var states []string
	added := make(map[string]bool)
	c := make(chan bool, 1)

	go func() {
		initState := strings.Repeat("0", len(b.FFs))
		statesToCheck := []string{initState}
		for {
			var state string
			state, statesToCheck = statesToCheck[0], statesToCheck[1:]
			reachable := b.reachableFromState(state)
			// Add all our new, reachable states
			for _, s := range reachable {
				// If we haven't seen it yet, add it to our list to check and note that
				// we've seen it
				if _, ok := added[s]; !ok {
					statesToCheck = append(statesToCheck, s)
					added[s] = true
				}
			}

			if len(statesToCheck) == 0 {
				break
			}
		}
		c <- true
	}()

	select {
	case <-c:
		// Results
	case <-time.After(time.Minute * 10):
		//Timeout
	}

	states = make([]string, len(added))
	i := 0
	for state := range added {
		states[i] = state
		i++
	}
	return states
}

// Given a state, returns a list of what states you can reach in 1-step
func (b *Bench) reachableFromState(state string) []string {
	// If there are n inputs, there are 2^n combinations of those inputs
	c := int(math.Pow(float64(2), float64(len(b.Inputs))))
	if states, ok := b.nextStates[state]; ok {
		return states
	}
	states := []string{}
	found := make(map[string]bool)
	for i := 0; i < c; i++ {
		// A bit mask padded with zeroes
		mask := fmt.Sprintf("%0"+strconv.Itoa(len(b.Inputs))+"b", i)
		b.resetPorts()
		b.setInputs(mask)
		b.setState(state)
		// Run the circuit
		b.run()
		// nextState is the state we've reached by running our sim
		nextState := b.State()
		// If we haven't seen this nextState yet
		if _, ok := found[nextState]; !ok {
			found[nextState] = true
			states = append(states, nextState)
		}
	}
	return states
}

// Run through a single step of the circuit
func (b *Bench) run() {
	gatesToCheck := []Gate{}
	gateCheck := make([]bool, len(b.Gates))
	gateIndex := 0

	// Our inputs are all ready
	for _, in := range b.Inputs {
		in.ready = true
		for _, conn := range in.conns {
			if conn.Type() != "DFF" && !gateCheck[conn.ID()-1] {
				gateCheck[conn.ID()-1] = true
				gatesToCheck = append(gatesToCheck, conn)
			}
		}
	}

	// Our state gates are all ready
	for _, g := range b.FFs {
		out := g.Outputs()[0]
		out.ready = true
		for _, conn := range out.conns {
			if conn.Type() != "DFF" && !gateCheck[conn.ID()-1] {
				gateCheck[conn.ID()-1] = true
				gatesToCheck = append(gatesToCheck, conn)
			}
		}
	}

	for {
		if inputsReady(gatesToCheck[gateIndex]) {
			// Pop off the gate
			var gate Gate
			gate, gatesToCheck = gatesToCheck[0], gatesToCheck[1:]
			gate.SetOut()
			gate.Outputs()[0].ready = true
			//debugStatement(fmt.Sprint("Looking at gate ID ", gate.ID(), " of type ", gate.Type(), " setting output to ", gate.Outputs()[0].on, " based on inputs ", gate.Inputs()[0].on), DEBUG)
			for _, conn := range gate.Outputs()[0].conns {
				// If  we have a new gate that hasn't been checked and isn't a DFF
				if conn.Type() != "DFF" && !gateCheck[conn.ID()-1] {
					gateCheck[conn.ID()-1] = true
					gatesToCheck = append(gatesToCheck, conn)
				}
			}
		}
		if len(gatesToCheck) > 0 {
			gateIndex = (gateIndex + 1) % len(gatesToCheck)
		} else {
			break
		}
	}
}

func (b *Bench) setInputs(mask string) {
	for i, bit := range mask {
		b.Inputs[i].on = bit == '1'
	}
}

func (b *Bench) setState(mask string) {
	for i, bit := range mask {
		b.FFs[i].Outputs()[0].on = bit == '1'
	}
}

func (b *Bench) setOutputs(mask string) {
	for i, bit := range mask {
		b.Outputs[i].on = bit == '1'
	}
}

func (b *Bench) resetPorts() {
	clearPorts(b.Inputs)
	clearPorts(b.Outputs)

	for _, gate := range b.Gates {
		clearPorts(gate.Inputs())
		clearPorts(gate.Outputs())
	}
}

func inputsReady(g Gate) bool {
	for _, in := range g.Inputs() {
		if !in.ready {
			return false
		}
	}
	return true
}

func clearPorts(ports []*Port) {
	for _, port := range ports {
		port.on = false
		port.ready = false
	}
}

func debugStatement(statement string, level int) {
	if level <= verbose {
		fmt.Println(statement)
	}
}
