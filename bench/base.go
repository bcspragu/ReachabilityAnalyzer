package bench

type Bench struct {
	// The state we're looking for
	Goal   string
	Unroll int

	// The bench file in a more convenient format
	lines []fileLine

	// Runners are used for finding valid state transitions
	runners []*runner

	/*
	 All of the fields below this block comment are various representational
	 aspects of the bench file, and are read by the runners, but don't need to be
	 modified, so we store a single copy of here, and give each runner a
	 reference to this bench
	*/
	// The ID fields are used to assign numeric values to gates and ports
	lastGateID int
	lastPortID int

	// Indexed by gate ID, each entry is a list of gates that are either inputs
	// or outputs of that gate
	toInputs  [][]int
	toOutputs [][]int

	// A mapping from port names to their IDs
	portMap map[string]int

	// Information about the bench file used for allocation
	gateCount  int
	inputCount int

	// Determines the output value of a gate based on its index
	gateType []gateType

	// Mapping from port name to the ID of the gate who's output it is
	gateOutputs map[string]int

	// Mapping from port name to a slice of gate IDs that take that port as an input
	gateInputs map[string][]int

	// List of FF gate IDs
	ffs []int

	// List of input gate IDs
	inputs []int

	// List of ports by gateID
	ports []ports
}

// The information from a single line in a bench file
type fileLine struct {
	output   string
	inputs   []string
	gateType string
	isIO     bool
}

// Port IDs from a single gate, for SAT solving
type ports struct {
	inputs []int
	output int
}

type State struct {
	state string
	input string
}

type newState struct {
	state string
	found State
}
