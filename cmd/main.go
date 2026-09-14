package main

import (
	"bufio"
	"fmt"
	"os"
	"tp-lox/internal/scanner"
)

type Lox struct{}

func newLox() *Lox {
	return &Lox{}
}

func (l *Lox) start() {
	args := os.Args

	if len(args) > 2 {
		fmt.Fprintln(os.Stderr, "usage: lox [script]")
		os.Exit(64)
	} else if len(args) == 2 {
		l.runFile(args[1])
	} else {
		l.runPrompt()
	}
}

func (l *Lox) runFile(path string) {
	fmt.Printf("-- Running in file mode: %s --\n", path)
}

func (l *Lox) runPrompt() {
	bfScanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")

	for bfScanner.Scan() {
		line := bfScanner.Text()
		l.run(line)
		fmt.Print("> ")
	}
}

func (l *Lox) run(source string) {
	scanner := scanner.NewScanner(source)
	tokens, err := scanner.Scan()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	for _, token := range tokens {
		fmt.Println(token)
	}
}

func main() {
	lox := newLox()
	lox.start()
}
