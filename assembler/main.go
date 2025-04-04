/*
The assembler translates assembly Hack programs (.asm files) into binary Hack programs (.hack
files).

Usage:

	assembler program.asm

This will read program.asm and write the binary program to program.hack.
*/
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/lfritz/nand2tetris/assembler/internal"
)

func main() {
	// parse command-line argument
	args := os.Args[1:]
	if len(args) != 1 {
		errorAndExit("Usage: assembler input.asm")
	}
	inPath := args[0]

	// figure out input and output file names
	if !strings.HasSuffix(inPath, ".asm") {
		errorAndExit("error: input filename must end in .asm")
	}
	outPath := strings.TrimSuffix(inPath, ".asm") + ".hack"

	// read input file
	inFile, err := os.Open(inPath)
	check(err)
	defer inFile.Close()

	// open output file
	outFile, err := os.Create(outPath)
	check(err)
	defer outFile.Close()

	// run the assembler
	err = internal.Run(inFile, outFile)
	check(err)
}

func check(err error) {
	if err == nil {
		return
	}
	errorAndExit("error: %v", err)
}

func errorAndExit(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}
