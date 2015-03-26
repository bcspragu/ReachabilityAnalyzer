package bench

type Dff struct {
	BaseGate
}

func (d *Dff) SetOut() {
	d.o.on = d.i.on
}

func NewDFF(id int, in, out *Port) Gate {
	dff := new(Dff)
	dff.id = id
	dff.gateType = "DFF"
	dff.attachPorts(in, out)

	return dff
}
