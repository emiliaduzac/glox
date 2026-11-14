package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
)

// sysexits.h error codes
const (
	EX_OK      = 0
	EX_USAGE   = 64 // The command was used incorrectly
	EX_DATAERR = 65 // The input data was incorrect in some way
	EX_NOINPUT = 66 // An input file (not a system file) did not exist or was not readable
)

const (
	ScanningMode = iota
	ParsingMode
	ResolveMode
	FullMode
	VMMode
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

// validateFlags verifica que la combinación de flags sea válida:
//   - scanning, parsing y resolve solo se usan solos (modo normal del tree-walking)
//   - vm puede ir solo o con debug
//   - debug requiere vm
func validateFlags(scanning, parsing, resolve, vm, debug bool) error {
	phases := 0
	for _, f := range []bool{scanning, parsing, resolve} {
		if f {
			phases++
		}
	}

	if phases > 1 {
		return errors.New("--scanning, --parsing and --resolve are mutually exclusive")
	}
	if phases == 1 && (vm || debug) {
		return errors.New("--scanning, --parsing and --resolve cannot be combined with --vm or --debug")
	}
	if debug && !vm {
		return errors.New("--debug can only be used with --vm")
	}
	return nil
}
