package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"tp-lox/internal/expression"
	"tp-lox/internal/parser"
	"tp-lox/internal/scanner"
)

// sysexits.h error codes
const (
	EX_OK      = 0
	EX_USAGE   = 64 // The command was used incorrectly
	EX_DATAERR = 65 // The input data was incorrect in some way
	EX_NOINPUT = 66 // An input file (not a system file) did not exist or was not readable
)

// manejo gracefull de SIGINT (Ctrl+C)
func gracefulShutdown(sigChan chan os.Signal, done chan struct{}) {
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		select {
		case <-sigChan:
			fmt.Println()
			os.Exit(EX_OK)
		case <-done:
			return
		}
	}()
}

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

	sigChan, done := make(chan os.Signal, 1), make(chan struct{})
	defer close(done)
	defer signal.Stop(sigChan)
	gracefulShutdown(sigChan, done)

	fmt.Print("> ")
	for bfScanner.Scan() {
		line := bfScanner.Text()
		_ = l.run(line)
		fmt.Print("> ")
	}
	fmt.Println()
}

func (l *Lox) run(source string) error {
	scanner := scanner.NewScanner(source)
	tokens, err := scanner.Scan()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	// for _, token := range tokens {
	// 	// por ahora siempre modo scanning
	// 	fmt.Println(token)
	// }

	parser := parser.NewParser(tokens)
	expr, err := parser.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	fmt.Println(expression.String(expr))
	return nil
}

func main() {
	lox := newLox()
	lox.start()
}
