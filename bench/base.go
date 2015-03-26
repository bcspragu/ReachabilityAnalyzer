package bench

import (
	"fmt"
)

type Port struct {
	id    int
	on    bool
	conns []Gate
}

// A BaseGate has a single input port, a single output port,
// and a method for getting that output
type BaseGate struct {
	// Our simple input and output ports
	i *Port
	o *Port

	gateType string
}

func (b *BaseGate) Out() bool {
	return b.o.on
}

func (b *BaseGate) Type() string {
	return b.gateType
}

// A utility method for debugging
func (b *BaseGate) Summary() string {
	return fmt.Sprint(b.Type(), " IN: ", b.i.id, " OUT: ", b.o.id)
}

func NewPort(id int) *Port {
	return &Port{id: id, conns: []Gate{}}
}

func (p *Port) attachGate(g Gate) error {
	p.conns = append(p.conns, g)
	return nil
}

func checkPortError(err error) {
	if err != nil {
		panic(err)
	}
}
