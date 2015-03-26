package bench

import (
	"fmt"
)

// An AND gate adds a second input port to a BaseGate, and
// sets its output based on both inputs
type And struct {
	BaseGate
	i2 *Port
}

func (a *And) SetOut() {
	a.o.on = a.i.on && a.i2.on
}

func (a *And) Inputs() []*Port {
	return []*Port{a.i, a.i2}
}

func NewAND(id int, in1, in2, out *Port) Gate {
	and := new(And)
	and.id = id
	and.gateType = "AND"

	and.i = in1
	and.i2 = in2

	and.o = out

	in1.attachGate(and)
	in2.attachGate(and)
	out.attachGate(and)

	return and
}

func (a *And) Summary() string {
	return fmt.Sprint("AND ", a.ID(), " IN1: ", a.i.id, " IN2: ", a.i2.id, " OUT: ", a.o.id)
}
