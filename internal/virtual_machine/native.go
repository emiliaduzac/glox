package virtualmachine

import (
	"bufio"
	"os"
	"time"
)

type nativeFn func(args []any) any

type native struct {
	arity int
	fn    nativeFn
}

func (n *native) String() string {
	return "<native fn>"
}

func builtinNatives() map[string]*native {
	return map[string]*native{
		"clock": {arity: 0, fn: clockNative},
		"input": {arity: 0, fn: inputNative},
	}
}

func clockNative(_ []any) any {
	return float64(time.Now().UnixNano()) / float64(time.Second)
}

var input = bufio.NewScanner(os.Stdin)

func inputNative(_ []any) any {
	if !input.Scan() {
		return nil
	}
	return input.Text()
}
