/*
The compiler translates Jack programs (.jack files) into Hack VM programs (.vm files). It can
compile either a single file or all .jack files in a directory.
Usage:

	compiler [flag] program.jack
	compiler [flag] directory

Flags:

	-t  instead of compiling, write tokens to an xml file
	-s  instead of compiling, write syntax tree to an xml file
*/
package main

import (
	"github.com/lfritz/nand2tetris/compiler/internal"

	"fmt"
	"os"
	"path"
	"strings"
)

type Mode int

const (
	ModeCompile Mode = iota
	ModePrintTokens
	ModePrintSyntax
)

func main() {
	// check command-line arguments
	args := os.Args[1:]
	mode := ModeCompile
	var inputPath string
	switch len(args) {
	case 1:
		inputPath = args[0]
	case 2:
		switch args[0] {
		case "-t":
			mode = ModePrintTokens
		case "-s":
			mode = ModePrintSyntax
		default:
			usageAndExit()
		}
		inputPath = args[1]
	default:
		usageAndExit()
	}

	info, err := os.Stat(inputPath)
	check(err)
	if info.IsDir() {
		compileDir(inputPath, mode)
	} else {
		compileFile(inputPath, mode)
	}
}

func compileDir(dirPath string, mode Mode) {
	dir, err := os.Open(dirPath)
	check(err)
	defer dir.Close()
	entries, err := dir.ReadDir(0)
	check(err)
	for _, info := range entries {
		filename := info.Name()
		if !info.IsDir() && strings.HasSuffix(filename, ".jack") {
			compileFile(path.Join(dirPath, filename), mode)
		}
	}
}

func compileFile(inPath string, mode Mode) {
	// figure out input and output file names
	filePath, ok := strings.CutSuffix(inPath, ".jack")
	if !ok {
		errorAndExit("error: input filename must end in .jack")
	}
	outputFiletype := "vm"
	if mode != ModeCompile {
		outputFiletype = "xml"
	}
	outPath := filePath + "." + outputFiletype

	// open input file
	inFile, err := os.Open(inPath)
	check(err)
	defer inFile.Close()

	// open output file
	outFile, err := os.Create(outPath)
	check(err)
	defer outFile.Close()

	// run the compiler
	filename := path.Base(filePath)
	switch mode {
	case ModeCompile:
		err = internal.Compile(filename, inFile, outFile)
	case ModePrintTokens:
		err = internal.PrintTokens(filename, inFile, outFile)
	case ModePrintSyntax:
		err = internal.PrintSyntax(filename, inFile, outFile)
	}
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

func usageAndExit() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "    compiler [flag] program.jack")
	fmt.Fprintln(os.Stderr, "    compiler [flag] directory")
	fmt.Fprintln(os.Stderr, "Flags:")
	fmt.Fprintln(os.Stderr, "    -t  instead of compiling, write tokens to an xml file")
	fmt.Fprintln(os.Stderr, "    -s  instead of compiling, write syntax tree to an xml file")
	os.Exit(1)
}
