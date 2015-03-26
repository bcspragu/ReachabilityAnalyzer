package bench

type Not struct {
	BaseGate
}

func (n *Not) SetOut() {
	n.o.on = !n.i.on
}

func NewNOT(in, out *Port) Gate {
	not := new(Not)
	not.gateType = "NOT"

	not.i = in
	not.o = out

	in.attachGate(not)
	out.attachGate(not)

	return not
}
