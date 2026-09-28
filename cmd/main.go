package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"tp-lox/internal/interpreter"
	"tp-lox/internal/parser"
	"tp-lox/internal/resolver"
	"tp-lox/internal/scanner"
	"tp-lox/internal/statements"
)

type Lox struct {
	mode        int
	interpreter *interpreter.Interpreter
}

func newLox(mode int) *Lox {
	return &Lox{
		mode:        mode,
		interpreter: interpreter.NewInterpreter(),
	}
}

func (l *Lox) start() {
	args := flag.Args()

	if len(args) > 1 {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd [--scanning|--parsing|--resolve] [script]")
		os.Exit(EX_USAGE)
	} else if len(args) == 1 {
		l.runFile(args[0])
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

	if l.mode == ScanningMode {
		for _, token := range tokens {
			fmt.Println(token)
		}
		return nil
	}

	parser := parser.NewParser(tokens)
	stmts, err := parser.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	if l.mode == ParsingMode {
		for _, stmt := range stmts {
			fmt.Println(statements.String(stmt))
		}
		return nil
	}

	resolver := resolver.NewResolver(l.interpreter)
	err = resolver.Resolve(stmts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	if l.mode == ResolveMode {
		fmt.Println("Interpreter Locals: " + interpreter.LocalsString(l.interpreter.Locals))
		return nil
	}

	err = l.interpreter.Interpret(stmts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

func main() {
	scanningFlag := flag.Bool("scanning", false, "corre solo el scanner, muestra los tokens")
	parsingFlag := flag.Bool("parsing", false, "corre el scanner y el parser, muestra el AST")
	resolveFlag := flag.Bool("resolve", false, "corre el scanner, el parser y el resolver, muestra las distancias resueltas de las variables locales")
	flag.Parse()

	var mode int
	switch {
	case *scanningFlag:
		mode = ScanningMode
	case *parsingFlag:
		mode = ParsingMode
	case *resolveFlag:
		mode = ResolveMode
	default:
		mode = FullMode
	}

	lox := newLox(mode)
	lox.start()
}
