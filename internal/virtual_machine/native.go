package virtualmachine

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
	"tp-lox/internal/utils"
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
		"str":   {arity: 1, fn: strNative},
		"num":   {arity: 1, fn: numNative},
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

func strNative(value []any) any {
	return utils.Stringify(value[0])
}

func numNative(args []any) any {
	text, ok := args[0].(string)
	if !ok {
		return nil
	}
	n, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil {
		return nil
	}
	return n
}
