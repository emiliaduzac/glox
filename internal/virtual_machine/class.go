package virtualmachine

type class struct {
	name string
}

func newClass(name string) *class {
	return &class{name: name}
}

func (c *class) String() string {
	return c.name
}
