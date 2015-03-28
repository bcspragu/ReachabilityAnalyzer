package bench

import (
	"bytes"
	//"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

type runner struct {
	id      int
	portMap map[string]*Port
	// NextStates is a map from a state string to all of the
	// states that can be reached from it, and what inputs are
	// needed to reach them

	lastPortID int
	lastGateID int

	Inputs  []*Port
	Outputs []*Port

	Gates []Gate
	FFs   []Gate
}

type NewState struct {
	State string
	Found State
}

type State struct {
	State string
	Input string
}

func newRunner(id int) *runner {
	r := &runner{id: id}
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

// Reads in states from inState channel, writes a list of what states you can reach in 1-step to foundStates
func (r *runner) reachableFromState(inStates <-chan string, foundStates chan<- NewState, searched chan<- bool) {
	for state := range inStates {
		debugStatement(fmt.Sprint("Runner ", r.id, " checking ", state), DEBUG)
		// If there are n inputs, there are 2^n combinations of those inputs
		c := int(math.Pow(float64(2), float64(len(r.Inputs))))

		// Keep track of the states we've found from here
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
				foundStates <- NewState{state, State{nextState, mask}}
				debugStatement(fmt.Sprint("Runner ", r.id, " found ", nextState), DEBUG)
			}
		}
		// Prevents a race condition between the foundStates and searched channels
		// which causes the program to exit before recording its last state
		// TODO(bsprague): Maybe have searched send completely searched states (as
		// strings) instead of info-less booleans
		time.Sleep(time.Millisecond)
		// Let the master know we've finished searching a state
		searched <- true
	}
	debugStatement(fmt.Sprint("Runner ", r.id, " finishing"), DEBUG)
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
