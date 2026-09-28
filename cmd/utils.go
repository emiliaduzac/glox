package main

import (
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
