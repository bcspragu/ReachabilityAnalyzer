package bench

type Not struct {
	BaseGate
}

func (n *Not) SetOut() {
	n.o.on = !n.i.on
}

func NewNOT(id int, in, out *Port) Gate {
	not := new(Not)
	not.id = id
	not.gateType = "NOT"
	not.attachPorts(in, out)

	return not
}

func (n *Not) attachPorts(in, out *Port) {
	n.i = in
	n.o = out

	in.attachGate(n)
	out.attachGate(n)
}
