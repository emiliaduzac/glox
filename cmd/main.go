package main

import (
	"bufio"
	"fmt"
	"os"
	"tp-lox/internal/scanner"
)

// sysexits.h error codes
const (
	EX_USAGE   = 64 // The command was used incorrectly
	EX_DATAERR = 65 // The input data was incorrect in some way
	EX_NOINPUT = 66 // An input file (not a system file) did not exist or was not readable
)

type Lox struct{}

func newLox() *Lox {
	return &Lox{}
}

func (l *Lox) start() {
	args := os.Args

	if len(args) > 2 {
		fmt.Fprintln(os.Stderr, "usage: lox [script]")
		os.Exit(EX_USAGE)
	} else if len(args) == 2 {
		l.runFile(args[1])
	} else {
		l.runPrompt()
	}
}

func (l *Lox) runFile(path string) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(EX_NOINPUT)
	}
	err = l.run(string(bytes))
	if err != nil {
		os.Exit(EX_DATAERR)
	}
}

func (l *Lox) runPrompt() {
	bfScanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")

	for bfScanner.Scan() {
		line := bfScanner.Text()
		_ = l.run(line)
		fmt.Print("> ")
	}
}

func (l *Lox) run(source string) error {
	scanner := scanner.NewScanner(source)
	tokens, err := scanner.Scan()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	for _, token := range tokens {
		// por ahora siempre modo scanning
		fmt.Println(token)
	}
	return nil
}

func main() {
	lox := newLox()
	lox.start()
}
