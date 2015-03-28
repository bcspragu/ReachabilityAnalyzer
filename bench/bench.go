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
var nRunner int
var verbose int

func init() {
	flag.IntVar(&verbose, "v", NONE, "level of verboseness in output")
	flag.IntVar(&nRunner, "runners", NONE, "How many ")
}

type Bench struct {
	runners []*runner
	// PortMap is a map from names in a bench file to their
	// corresponding port objects
	nextStates map[string][]string

	Goal string
}

func NewFromFile(filename string) (*Bench, error) {
	debugStatement("Creating bench", VERBOSE)
	bench := new(Bench)
	bench.nextStates = make(map[string][]string)
	bench.runners = make([]*runner, nRunner)

	goalState, err := ioutil.ReadFile(filename + ".state")
	if err != nil {
		return bench, err
	}
	bench.Goal = strings.TrimSpace(string(goalState))

	for i := 0; i < nRunner; i++ {
		bench.runners[i] = newRunner()
	}

	file, err := os.Open(filename + ".bench")
	if err != nil {
		return nil, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		for i := 0; i < nRunner; i++ {
			bench.runners[i].ParseLine(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return bench, nil
}

func newRunner() *runner {
	r := new(runner)
	r.Gates = []Gate{}
	r.FFs = []Gate{}
	r.portMap = make(map[string]*Port)

	r.Inputs = []*Port{}
	r.Outputs = []*Port{}
	return r
}

func (r *runner) ParseLine(line string) {
	matches := gateRE.FindStringSubmatch(line)
	// If we have matches, it's a gate statement
	if len(matches) > 0 {
		out, gate := matches[1], matches[2]
		switch gate {
		case "AND":
			r.AddAND(matches[3], matches[4], out)
		case "NOT":
			r.AddNOT(matches[3], out)
		case "DFF":
			r.AddDFF(matches[3], out)
		}
	} else {
		// Otherwise, it's an input or output
		if inOutRE.MatchString(line) {
			ioMatch := inOutRE.FindStringSubmatch(line)
			io, port := ioMatch[1], ioMatch[2]
			switch io {
			case "INPUT":
				r.AddInput(port)
			case "OUTPUT":
				r.AddOutput(port)
			}
		}
	}
}

func (r *runner) AddAND(in1, in2, out string) {
	inPort1 := r.FindOrCreatePort(in1)
	inPort2 := r.FindOrCreatePort(in2)
	outPort := r.FindOrCreatePort(out)

	and := NewAND(r.nextGateID(), inPort1, inPort2, outPort)
	r.Gates = append(r.Gates, and)
}

func (r *runner) AddNOT(in, out string) {
	r.addOneInputGate(in, out, NewNOT)
}

func (r *runner) AddDFF(in, out string) {
	gate := r.addOneInputGate(in, out, NewDFF)
	r.FFs = append(r.FFs, gate)
}

func (r *runner) AddInput(p string) {
	r.Inputs = append(r.Inputs, r.FindOrCreatePort(p))
}

func (r *runner) AddOutput(p string) {
	r.Outputs = append(r.Outputs, r.FindOrCreatePort(p))
}

func (r *runner) addOneInputGate(in, out string, newGate func(int, *Port, *Port) Gate) Gate {
	inPort := r.FindOrCreatePort(in)
	outPort := r.FindOrCreatePort(out)

	gate := newGate(r.nextGateID(), inPort, outPort)
	r.Gates = append(r.Gates, gate)
	return gate
}

func (r *runner) FindOrCreatePort(name string) *Port {
	if port, ok := r.portMap[name]; ok {
		// Port exists
		return port
	}
	// Create the port if it doesn't exist, using the next
	// available sequential id
	port := NewPort(r.nextPortID())
	r.portMap[name] = port
	return port
}

func (r *runner) nextPortID() int {
	r.lastPortID++
	return r.lastPortID
}

func (r *runner) nextGateID() int {
	r.lastGateID++
	return r.lastGateID
}

func (r *runner) State() string {
	// One character for each flip flop
	buf := make([]byte, 0, len(r.FFs))
	buffer := bytes.NewBuffer(buf)

	for _, dff := range r.FFs {
		if dff.Inputs()[0].on {
			buffer.WriteString("1")
		} else {
			buffer.WriteString("0")
		}
	}

	return buffer.String()
}

func (r *runner) Summary() string {
	var buffer bytes.Buffer

	for _, i := range r.Inputs {
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

	for _, o := range r.Outputs {
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

	for _, gate := range r.Gates {
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

func (r *runner) ReachableStates() []string {
	var states []string
	added := make(map[string]bool)
	c := make(chan bool, 1)

	go func() {
		initState := strings.Repeat("0", len(r.FFs))
		statesToCheck := []string{initState}
		for {
			var state string
			state, statesToCheck = statesToCheck[0], statesToCheck[1:]
			reachable := r.reachableFromState(state)
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
func (r *runner) reachableFromState(state string) []string {
	// If there are n inputs, there are 2^n combinations of those inputs
	c := int(math.Pow(float64(2), float64(len(r.Inputs))))
	if states, ok := r.b.nextStates[state]; ok {
		return states
	}
	states := []string{}
	found := make(map[string]bool)
	for i := 0; i < c; i++ {
		// A bit mask padded with zeroes
		mask := fmt.Sprintf("%0"+strconv.Itoa(len(r.Inputs))+"b", i)
		r.resetPorts()
		r.setInputs(mask)
		r.setState(state)
		// Run the circuit
		r.run()
		// nextState is the state we've reached by running our sim
		nextState := r.State()
		// If we haven't seen this nextState yet
		if _, ok := found[nextState]; !ok {
			found[nextState] = true
			states = append(states, nextState)
		}
	}
	return states
}

// Run through a single step of the circuit
func (r *runner) run() {
	gatesToCheck := []Gate{}
	gateCheck := make([]bool, len(r.Gates))
	gateIndex := 0

	// Our inputs are all ready
	for _, in := range r.Inputs {
		in.ready = true
		for _, conn := range in.conns {
			if conn.Type() != "DFF" && !gateCheck[conn.ID()-1] {
				gateCheck[conn.ID()-1] = true
				gatesToCheck = append(gatesToCheck, conn)
			}
		}
	}

	// Our state gates are all ready
	for _, g := range r.FFs {
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

func (r *runner) setInputs(mask string) {
	for i, bit := range mask {
		r.Inputs[i].on = bit == '1'
	}
}

func (r *runner) setState(mask string) {
	for i, bit := range mask {
		r.FFs[i].Outputs()[0].on = bit == '1'
	}
}

func (r *runner) setOutputs(mask string) {
	for i, bit := range mask {
		r.Outputs[i].on = bit == '1'
	}
}

func (r *runner) resetPorts() {
	clearPorts(r.Inputs)
	clearPorts(r.Outputs)

	for _, gate := range r.Gates {
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
