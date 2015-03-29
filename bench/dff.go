package bench

type Dff struct {
	BaseGate
}

func (d *Dff) SetOut() {
	// NOP
}

func NewDFF(id int, in, out *Port) Gate {
	dff := new(Dff)
	dff.id = id
	dff.gateType = "DFF"
	dff.attachPorts(in, out)

	return dff
}

func (d *Dff) attachPorts(in, out *Port) {
	d.i = in
	d.o = out

	in.attachGate(d)
	out.attachGate(d)
}
