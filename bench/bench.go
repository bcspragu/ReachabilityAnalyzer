package bench

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
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
var gateRE = regexp.MustCompile(`^(\w+)\s*=\s*(\w+)\((\w+)\s*(?:,\s*(\w+))?\)$`)
var inOutRE = regexp.MustCompile(`^(\w+)\((\w+)\)$`)

const (
	None = iota
	Debug
	Verbose
)

func NewFromFile(filename string, nRunners int) (*Bench, error) {
	bench := &Bench{RunnerCount: nRunners}
	bench.runners = make([]*runner, bench.RunnerCount)

	bench.portMap = make(map[string]int)
	bench.gateInputs = make(map[string][]int)
	bench.gateOutputs = make(map[string]int)

	goalState, err := ioutil.ReadFile(filename + ".state")
	if err != nil {
		return nil, err
	}
	bench.Goal = strings.TrimSpace(string(goalState))

	file, err := os.Open(filename + ".bench")
	if err != nil {
		return nil, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		bench.lines = append(bench.lines, toFileLine(line))
	}

	bench.loadLines(bench.lines)
	bench.parseLines(bench.lines)
	for i := 0; i < bench.RunnerCount; i++ {
		bench.runners[i] = &runner{id: i, outState: make([]outState, len(bench.toOutputs)), b: bench}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return bench, nil
}

func (b *Bench) loadLines(f []fileLine) {
	var id int
	for _, line := range f {

		// Inputs and gates get IDs
		if !line.isIO || line.gateType == "INPUT" {
			id = b.nextGateID()
		}

		if !line.isIO {
			b.gateCount++
		}

		switch line.gateType {
		case "INPUT":
			b.inputCount++
			b.gateOutputs[line.output] = id
		case "OUTPUT":
			// Currently we don't do anything with outputs
		case "AND":
			b.gateOutputs[line.output] = id
			b.gateInputs[line.inputs[0]] = append(b.gateInputs[line.inputs[0]], id)
			b.gateInputs[line.inputs[1]] = append(b.gateInputs[line.inputs[1]], id)
		case "NOT", "DFF":
			b.gateOutputs[line.output] = id
			b.gateInputs[line.inputs[0]] = append(b.gateInputs[line.inputs[0]], id)
		}
	}

	totalCount := b.gateCount + b.inputCount

	b.gateType = make([]gateType, totalCount)
	b.toInputs = make([][]int, totalCount)
	b.toOutputs = make([][]int, totalCount)
	b.ports = make([]ports, totalCount)

	b.lastGateID = 0
}

func (b *Bench) parseLines(f []fileLine) {
	for _, line := range f {
		switch line.gateType {
		case "AND":
			b.addAND(line.inputs[0], line.inputs[1], line.output)
		case "NOT":
			b.addNOT(line.inputs[0], line.output)
		case "DFF":
			b.addDFF(line.inputs[0], line.output)
		case "INPUT":
			b.addInput(line.output)
		case "OUTPUT":
			// We don't do anything with it
		}
	}
}

func (b *Bench) addAND(in1, in2, out string) {
	id := b.nextGateID()
	b.gateType[id].and = true

	// The gate whose output is attached to in1 gets this id added to its
	// outputs, and we add that gate to our list of inputs
	g := b.gateOutputs[in1]
	//b.toOutputs[g].to = append(b.toOutputs[g].to, id)
	b.toInputs[id] = append(b.toInputs[id], g)

	g = b.gateOutputs[in2]
	//b.toOutputs[g].to = append(b.toOutputs[g].to, id)
	b.toInputs[id] = append(b.toInputs[id], g)

	// We find a list of all the gates that use our output as an input
	gs := b.gateInputs[out]
	for _, g := range gs {
		//b.toInputs[g] = append(b.toInputs[g], id)
		b.toOutputs[id] = append(b.toOutputs[id], g)
	}
	b.ports[id] = ports{inputs: []int{b.portID(in1), b.portID(in2)}, output: b.portID(out)}
}

func (b *Bench) addNOT(in, out string) {
	id := b.addOneInputGate(in, out)
	b.gateType[id].not = true
}

func (b *Bench) addDFF(in, out string) {
	id := b.addOneInputGate(in, out)
	b.ffs = append(b.ffs, id)
	b.gateType[id].ff = true
}

func (b *Bench) addInput(out string) {
	id := b.nextGateID()
	b.gateType[id].input = true

	// We find a list of all the gates that use our output as an input
	gs := b.gateInputs[out]
	for _, g := range gs {
		//b.toInputs[g] = append(b.toInputs[g], id)
		b.toOutputs[id] = append(b.toOutputs[id], g)
	}

	b.inputs = append(b.inputs, id)
}

func (b *Bench) addOneInputGate(in, out string) int {
	id := b.nextGateID()

	// The gate whose output is attached to in1 gets this id added to its
	// outputs, and we add that gate to our list of inputs
	g := b.gateOutputs[in]
	//b.toOutputs[g].to = append(b.toOutputs[g].to, id)
	b.toInputs[id] = append(b.toInputs[id], g)

	// We find a list of all the gates that use our output as an input
	gs := b.gateInputs[out]
	for _, g := range gs {
		//b.toInputs[g] = append(b.toInputs[g], id)
		b.toOutputs[id] = append(b.toOutputs[id], g)
	}
	b.ports[id] = ports{inputs: []int{b.portID(in)}, output: b.portID(out)}
	return id
}

// Return new IDs, starting at zero
func (b *Bench) nextGateID() int {
	b.lastGateID++
	return b.lastGateID - 1
}

// Return new IDs, starting at one
func (b *Bench) nextPortID() int {
	b.lastPortID++
	return b.lastPortID
}

func (b *Bench) portID(name string) int {
	if id, ok := b.portMap[name]; ok {
		return id
	} else {
		id := b.nextPortID()
		b.portMap[name] = id
		return id
	}
}

func toFileLine(line string) fileLine {
	fileLine := fileLine{}
	matches := gateRE.FindStringSubmatch(line)
	// If we have matches, it's a gate statement
	if gateRE.MatchString(line) {
		out, gate := matches[1], matches[2]
		fileLine.gateType = gate
		fileLine.output = out
		switch gate {
		case "AND":
			fileLine.inputs = matches[3:5]
		case "NOT", "DFF":
			fileLine.inputs = matches[3:4]
		}
	} else if inOutRE.MatchString(line) {
		ioMatch := inOutRE.FindStringSubmatch(line)
		io, port := ioMatch[1], ioMatch[2]

		fileLine.isIO = true
		fileLine.gateType = io
		fileLine.output = port
	}

	return fileLine
}

func (b *Bench) ReachableStates() map[string][]State {
	return b.reachableStates(func(s string) bool {
		return false
	})
}

func (b *Bench) IsReachable() (bool, map[string][]State) {
	states := b.reachableStates(func(s string) bool {
		return s == b.Goal
	})

	for state := range states {
		if state == b.Goal {
			return true, states
		}
	}
	return false, states
}

// To find all of the reachable states, we spin up a bunch of worker threads.
// Every time a worker thread finds a new state, it passes it back over the
// channel, and we place it onto the queue for another worker to use
func (b *Bench) reachableStates(goalFunc func(string) bool) map[string][]State {
	initState := strings.Repeat("0", len(b.ffs))
	nextStates := make(map[string][]State)
	nextStates[initState] = []State{}

	statesToCheck := make(chan string, 1000)
	foundStates := make(chan newState, 1000)
	searched := make(chan bool, b.RunnerCount)
	statesToCheck <- initState

	// Spin up our runners
	for _, r := range b.runners {
		go r.reachableFromState(statesToCheck, foundStates, searched)
	}

	totalStates := 1
Loop:
	for {
		select {
		case found := <-foundStates:
			potentiallyNewState := found.found.state
			// If it's actually new
			if _, ok := nextStates[potentiallyNewState]; !ok {
				totalStates++
				statesToCheck <- potentiallyNewState
				nextStates[found.found.state] = []State{}
				nextStates[found.state] = append(nextStates[found.state], found.found)
				b.debugStatement(fmt.Sprint("Sent ", potentiallyNewState, " to be searched"), Debug)
				if goalFunc(potentiallyNewState) {
					break Loop
				}
			}
		case <-searched: // A state has been finished
			totalStates--
			b.debugStatement(fmt.Sprint(totalStates, " left"), Debug)
			if totalStates == 0 {
				close(statesToCheck)
				break Loop
			}
		case <-time.After(time.Minute * 10):
			b.debugStatement("Timed out", Debug)
			break Loop
		}
	}

	return nextStates
}

func (b *Bench) debugStatement(statement string, level int) {
	if level <= b.LogLevel {
		fmt.Println(statement)
	}
}

func (b *Bench) Solution(nextStates map[string][]State) string {
	prevStates := make(map[string][]State)
	for state, states := range nextStates {
		for _, prevState := range states {
			prevStates[prevState.state] = append(prevStates[prevState.state], State{state: state, input: prevState.input})
		}
	}
	currState := b.Goal
	initState := strings.Repeat("0", len(b.ffs))
	strs := []string{fmt.Sprint("Final: ", b.Goal, "\n")}
	i := 0
Loop:
	for {
		for _, prevState := range prevStates[currState] {
			if prevState.state == initState {
				strs = append(strs, fmt.Sprint("Initial: ", initState, " Inputs: ", prevState.input, "\n"))
				break Loop
			}
		}
		prev := prevStates[currState][0]
		strs = append(strs, fmt.Sprint(": ", prev.state, " Inputs: ", prev.input, "\n"))
		currState = prev.state
		i++
	}

	var buf bytes.Buffer
	for i := len(strs) - 1; i >= 0; i-- {
		if i != 0 && i != len(strs)-1 {
			buf.WriteString("State " + strconv.Itoa(len(strs)-i) + strs[i])
		} else {
			buf.WriteString(strs[i])
		}
	}
	return buf.String()
}

func (b *Bench) NextState(state, input string) string {
	r := b.runners[0]
	r.clearState()
	r.setInputs(input)
	r.setState(state)
	// Run the circuit
	r.run()
	// nextState is the state we've reached by running our sim
	return r.State()
}
