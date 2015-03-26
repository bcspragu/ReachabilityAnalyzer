package bench

type Dff struct {
	BaseGate
}

func (d *Dff) SetOut() {
	d.o.on = d.i.on
}

func NewDFF(in, out *Port) Gate {
	dff := new(Dff)
	dff.gateType = "DFF"

	dff.i = in
	dff.o = out

	in.attachGate(dff)
	out.attachGate(dff)

	return dff
}
