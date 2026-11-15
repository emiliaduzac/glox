package virtualmachine

import c "tp-lox/internal/chunk"

type closure struct {
	function *c.Function
}

func newClosure(function *c.Function) *closure {
	return &closure{function: function}
}

func (cl *closure) String() string {
	return cl.function.String()
}
