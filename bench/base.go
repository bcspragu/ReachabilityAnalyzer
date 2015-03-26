package bench

import (
	"fmt"
)

type Port struct {
	id    int
	on    bool
	ready bool
	conns []Gate
}

type Gate interface {
	// Set the output port value
	SetOut()

	// A slice of all the associated ports
	Inputs() []*Port
	Outputs() []*Port

	// Get the type of the gate (AND, NOT, DFF)
	Type() string

	// Gates have IDs like ports, but Gate IDs have no relation
	// to port IDs
	ID() int
}

// A BaseGate has a single input port, a single output port,
// and a method for getting that output
type BaseGate struct {
	id int

	// Our simple input and output ports
	i *Port
	o *Port

	gateType string
}

func (b *BaseGate) Inputs() []*Port {
	return []*Port{b.i}
}

func (b *BaseGate) Outputs() []*Port {
	return []*Port{b.o}
}

func (b *BaseGate) Type() string {
	return b.gateType
}

// A utility method for debugging
func (b *BaseGate) Summary() string {
	return fmt.Sprint(b.Type(), " ", b.ID(), " IN: ", b.i.id, " OUT: ", b.o.id)
}

func NewPort(id int) *Port {
	return &Port{id: id, conns: []Gate{}}
}

func (p *Port) attachGate(g Gate) {
	p.conns = append(p.conns, g)
}

func (p *Port) ID() int {
	return p.id
}

func (b *BaseGate) ID() int {
	return b.id
}

func (p *Port) Conns() []Gate {
	return p.conns
}
