package virtualmachine

type nativeFn func(args []any) any

type native struct {
	arity int
	fn    nativeFn
}

func (n *native) String() string {
	return "<native fn>"
}
