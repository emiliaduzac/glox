package chunk

import "fmt"

type Function struct {
	Arity        int
	Chunk        Chunk
	Name         string
	UpvalueCount int
}

func NewFunction() *Function {
	return &Function{}
}

func (f *Function) String() string {
	if f.Name == "" {
		return "<script>"
	}
	return fmt.Sprintf("<fn %s>", f.Name)
}
