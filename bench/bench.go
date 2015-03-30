package bench

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
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
	NONE = iota
	DEBUG
	VERBOSE
)

var nRunner int
var verbose int
var unroll int

func init() {
	flag.IntVar(&verbose, "v", NONE, "level of verboseness in output")
	flag.IntVar(&nRunner, "runners", 10, "how many processes to simultaneously calculate state")
	flag.IntVar(&unroll, "unroll", 2, "how many times to unroll the formula")
}

type Bench struct {
	runners []*runner
	// PortMap is a map from names in a bench file to their
	// corresponding port objects
	nextStates map[string][]State

	Goal   string
	unroll int
}

func NewFromFile(filename string) (*Bench, error) {
	debugStatement("Creating bench", VERBOSE)
	bench := &Bench{unroll: unroll}
	bench.nextStates = make(map[string][]State)
	bench.runners = make([]*runner, nRunner)

	goalState, err := ioutil.ReadFile(filename + ".state")
	if err != nil {
		return nil, err
	}
	bench.Goal = strings.TrimSpace(string(goalState))

	for i := 0; i < nRunner; i++ {
		bench.runners[i] = newRunner(i)
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
			// TODO(bsprague): Do this more efficiently, so that you don't need to
			// parse the same line nRunners times
			bench.runners[i].count(line)
		}
	}

	for _, r := range bench.runners {
		r.setSize()
	}

	file.Seek(0, 0)
	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		for i := 0; i < nRunner; i++ {
			// TODO(bsprague): Do this more efficiently, so that you don't need to
			// parse the same line nRunners times
			bench.runners[i].ParseLine(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return bench, nil
}

func (b *Bench) ReachableStates() []string {
	return b.reachableStates(func(s string) bool {
		return false
	})
}

func (b *Bench) IsReachable() bool {
	states := b.reachableStates(func(s string) bool {
		return s == b.Goal
	})

	for _, state := range states {
		if state == b.Goal {
			return true
		}
	}
	return false
}

// To find all of the reachable states, we spin up a bunch of worker threads.
// Every time a worker thread finds a new state, it passes it back over the
// channel, and we place it onto the queue for another worker to use
func (b *Bench) reachableStates(goalFunc func(string) bool) []string {
	var states []string

	// TODO(bsprague): Record basic statistics about the bench file so that we
	// don't have to rely on runners for this info
	initState := strings.Repeat("0", len(b.runners[0].FFs))
	b.nextStates[initState] = []State{}

	statesToCheck := make(chan string, 1000)
	foundStates := make(chan NewState, 1000)
	searched := make(chan bool, nRunner)
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
			potentiallyNewState := found.Found.State
			// If it's actually new
			if _, ok := b.nextStates[potentiallyNewState]; !ok {
				totalStates++
				statesToCheck <- potentiallyNewState
				b.nextStates[found.Found.State] = []State{}
				b.nextStates[found.State] = append(b.nextStates[found.State], found.Found)
				debugStatement(fmt.Sprint("Sent ", potentiallyNewState, " to be searched"), DEBUG)
				if goalFunc(potentiallyNewState) {
					break Loop
				}
			}
		case <-searched: // A state has been finished
			totalStates--
			debugStatement(fmt.Sprint(totalStates, " left"), DEBUG)
			if totalStates == 0 {
				close(statesToCheck)
				break Loop
			}
		case <-time.After(time.Minute * 10):
			debugStatement("Timed out", DEBUG)
			break Loop
		}
	}

	states = make([]string, len(b.nextStates))
	i := 0
	for state := range b.nextStates {
		states[i] = state
		i++
	}
	return states
}

func debugStatement(statement string, level int) {
	if level <= verbose {
		fmt.Println(statement)
	}
}

func (b *Bench) PortMap() string {
	var buf bytes.Buffer
	for name, port := range b.runners[0].portMap {
		buf.WriteString(fmt.Sprintln(name, "-", port.id))
	}
	return buf.String()
}

func (b *Bench) SetUnroll(unroll int) {
	b.unroll = unroll
}

func (b *Bench) NextState(state, input string) string {
	r := b.runners[0]
	r.resetPorts()
	r.setInputs(input)
	r.setState(state)
	// Run the circuit
	r.run()
	// nextState is the state we've reached by running our sim
	return r.State()
}
