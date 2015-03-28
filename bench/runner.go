package bench

type runner struct {
	b *Bench

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

type State struct {
	State string
	Input string
}
