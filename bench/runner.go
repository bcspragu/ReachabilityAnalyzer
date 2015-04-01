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
	id int
	b  *Bench

	outState []outState
}

type outState struct {
	ready bool
	on    bool
}

func (r *runner) State() string {
	// One character for each flip flop
	buf := make([]byte, 0, len(r.b.ffs))
	buffer := bytes.NewBuffer(buf)

	for _, id := range r.b.ffs {
		if r.outState[r.b.toInputs[id][0]].on {
			buffer.WriteString("1")
		} else {
			buffer.WriteString("0")
		}
	}

	return buffer.String()
}

// Reads in states from inState channel, writes a list of what states you can reach in 1-step to foundStates
func (r *runner) reachableFromState(inStates <-chan string, foundStates chan<- newState, searched chan<- bool) {
	for state := range inStates {
		r.b.debugStatement(fmt.Sprint("Runner ", r.id, " checking ", state), Debug)
		// If there are n inputs, there are 2^n combinations of those inputs
		c := int(math.Pow(float64(2), float64(r.b.inputCount)))

		// Keep track of the states we've found from here
		found := make(map[string]bool)

		for i := 0; i < c; i++ {
			// A bit mask padded with zeroes
			mask := fmt.Sprintf("%0"+strconv.Itoa(r.b.inputCount)+"b", i)
			r.clearState()
			r.setInputs(mask)
			r.setState(state)
			// Run the circuit
			r.run()
			// nextState is the state we've reached by running our sim
			nextState := r.State()
			// If we haven't seen this nextState yet
			if _, ok := found[nextState]; !ok {
				found[nextState] = true
				foundStates <- newState{state, State{nextState, mask}}
				r.b.debugStatement(fmt.Sprint("Runner ", r.id, " found ", nextState), Debug)
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
	r.b.debugStatement(fmt.Sprint("Runner ", r.id, " finishing"), Debug)
}

// Run through a single step of the circuit
func (r *runner) run() {
	b := r.b
	gatesToCheck := []int{}
	gateCheck := make([]bool, b.gateCount+b.inputCount)
	// Our inputs are all ready
	for _, in := range b.inputs {
		r.outState[in].ready = true
		for _, id := range b.toOutputs[in] {
			if !b.gateType[id].ff && !gateCheck[id] {
				if r.inputsReady(id) {
					gateCheck[id] = true
					gatesToCheck = append(gatesToCheck, id)
				}
			}
		}
	}

	// Our state gates are all ready
	for _, g := range b.ffs {
		r.outState[g].ready = true
		for _, id := range b.toOutputs[g] {
			if !b.gateType[id].ff && !gateCheck[id] {
				if r.inputsReady(id) {
					gateCheck[id] = true
					gatesToCheck = append(gatesToCheck, id)
				}
			}
		}
	}

	for {
		// Pop off the gate
		var gate int
		gate, gatesToCheck = gatesToCheck[0], gatesToCheck[1:]
		var on bool
		if b.gateType[gate].and {
			on = r.isOnAND(gate)
		} else if b.gateType[gate].not {
			on = r.isOnNOT(gate)
		}
		out := r.outState[gate]

		out.on = on
		out.ready = true
		r.outState[gate] = out
		for _, id := range b.toOutputs[gate] {
			// If  we have a new gate that hasn't been checked
			if !b.gateType[id].ff && !gateCheck[id] {
				if r.inputsReady(id) {
					gateCheck[id] = true
					gatesToCheck = append(gatesToCheck, id)
				}
			}
		}
		if len(gatesToCheck) == 0 {
			break
		}
	}
}

func (r *runner) setInputs(mask string) {
	for i, bit := range mask {
		r.outState[r.b.inputs[i]].on = bit == '1'
	}
}

func (r *runner) clearState() {
	for i := range r.outState {
		r.outState[i].ready = false
	}
}

func (r *runner) setState(mask string) {
	for i, bit := range mask {
		r.outState[r.b.ffs[i]].on = bit == '1'
	}
}

func (r *runner) inputsReady(g int) bool {
	for _, in := range r.b.toInputs[g] {
		// in is the id of the gate who's output is connected to one of g's inputs
		if !r.outState[in].ready {
			return false
		}
	}
	return true
}

func (r *runner) isOnAND(g int) bool {
	for _, in := range r.b.toInputs[g] {
		// in is the id of the gate who's output is connected to one of g's inputs
		if !r.outState[in].on {
			return false
		}
	}
	return true
}

func (r *runner) isOnNOT(g int) bool {
	in := r.b.toInputs[g][0]
	return !r.outState[in].on
}
