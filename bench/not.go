package bench

type Not struct {
	i Port
	o Port
}

func (n *Not) SetOut() {
	n.o.on = !n.i.on
}

func (n *Not) Out() bool {
	return n.o.on
}
