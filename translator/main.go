/*
The translator translates Hack VM programs (.vm files) into Hack assembly programs (.asm files).

Usage:

	translator program.vm

This will read program.vm and write assembly code to program.asm.
*/
package main

import (
	"github.com/lfritz/nand2tetris/translator/internal"

	"fmt"
	"os"
	"path"
	"strings"
)

func main() {
	// check command-line argument
	args := os.Args[1:]
	if len(args) != 1 {
		errorAndExit("Usage: translator input.vm")
	}

	// figure out input and output file names
	inPath := args[0]
	filename, ok := strings.CutSuffix(inPath, ".vm")
	if !ok {
		errorAndExit("error: input filename must end in .vm")
	}
	outPath := filename + ".asm"

	// open input file
	inFile, err := os.Open(inPath)
	check(err)
	defer inFile.Close()

	// open output file
	outFile, err := os.Create(outPath)
	check(err)
	defer outFile.Close()

	// run the translator
	err = internal.Run(path.Base(filename), inFile, outFile)
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
