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

type instance struct {
	class  *class
	fields map[string]any
}

func newInstance(class *class) *instance {
	return &instance{class: class, fields: map[string]any{}}
}

func (i *instance) String() string {
	return i.class.name + " instance"
}
