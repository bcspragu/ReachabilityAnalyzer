package bench

type Port struct {
	id int
	on bool
}

type And struct {
	i1 Port
	i2 Port
	o  Port
}

func (a *And) SetOut() {
	a.o.on = a.i1.on && a.i2.on
}

func (a *And) Out() bool {
	return a.o.on
}
