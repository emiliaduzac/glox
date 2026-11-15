package virtualmachine

import c "tp-lox/internal/chunk"

type upvalue struct {
	location *any
}

func newUpvalue(slot *any) *upvalue {
	return &upvalue{location: slot}
}

type closure struct {
	function *c.Function
	upvalues []*upvalue
}

func newClosure(function *c.Function) *closure {
	return &closure{function: function, upvalues: make([]*upvalue, function.UpvalueCount)}
}

func (cl *closure) String() string {
	return cl.function.String()
}
