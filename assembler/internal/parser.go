package internal

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// Parser implements a parser for Hack assembly code.
type Parser struct {
	scanner *bufio.Scanner
	current string
}

// NewParser creates a new parser given the contents of a Hack assembly file.
//
// To use the Parser, call Scan to get the next line, then InstructionType to get the type of the
// current instruction, then the method for that type to get the actual instruction.
func NewParser(r io.Reader) *Parser {
	p := Parser{
		scanner: bufio.NewScanner(r),
	}
	return &p
}

// Scan advances the parser to the next line of assembly code, skipping empty lines and comments.
// It returns false when the end of the source has been reached.
func (p *Parser) Scan() bool {
	for p.scanner.Scan() {
		line := p.scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "//") {
			continue
		}
		p.current = line
		return true
	}
	return false
}

// InstructionType returns the type of the current instruction. This is only valid after Scan has
// been called and returned true.
func (p *Parser) InstructionType() InstructionType {
	return instructionType(p.current)
}

// SymbolicAInstruction parses and returns a symbolic A-instruction. Only valid if InstructionType
// returns TypeASymbolic.
func (p *Parser) SymbolicAInstruction() (SymbolicAInstruction, error) {
	return parseSymbolicAInstruction(p.current)
}

// SymbolicAInstruction parses and returns a decimal A-instruction. Only valid if InstructionType
// returns TypeADecimal.
func (p *Parser) DecimalAInstruction() (DecimalAInstruction, error) {
	return parseDecimalAInstruction(p.current)
}

// SymbolicAInstruction parses and returns a C-instruction. Only valid if InstructionType returns
// TypeC.
func (p *Parser) CInstruction() (CInstruction, error) {
	return parseCInstruction(p.current)
}

// SymbolicAInstruction parses and returns a label pseudo-instruction. Only valid if InstructionType
// returns TypeL.
func (p *Parser) LInstruction() (LInstruction, error) {
	return parseLInstruction(p.current)
}

func instructionType(line string) InstructionType {
	if remaining, ok := strings.CutPrefix(line, "@"); ok {
		if validSymbol(remaining) {
			return TypeASymbolic
		} else {
			return TypeADecimal
		}
	}
	if strings.HasPrefix(line, "(") {
		return TypeL
	}
	return TypeC
}

func validSymbol(symbol string) bool {
	if len(symbol) == 0 {
		return false
	}
	for index, c := range symbol {
		switch {
		case unicode.IsLetter(c):
		case c == '_' || c == '.' || c == '$' || c == ':':
		case unicode.IsDigit(c):
			if index == 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func parseSymbolicAInstruction(line string) (instruction SymbolicAInstruction, err error) {
	symbol, ok := strings.CutPrefix(line, "@")
	if !ok {
		err = fmt.Errorf("invalid A-instruction: '%s'", line)
		return
	}
	instruction = SymbolicAInstruction{
		Symbol: symbol,
	}
	return
}

func parseDecimalAInstruction(line string) (instruction DecimalAInstruction, err error) {
	number, ok := strings.CutPrefix(line, "@")
	if !ok {
		err = fmt.Errorf("invalid A-instruction: '%s'", line)
		return
	}
	var value uint64
	value, err = strconv.ParseUint(number, 10, 15)
	if err != nil {
		err = fmt.Errorf("invalid A-instruction: '%s'", line)
		return
	}
	instruction = DecimalAInstruction{
		Value: uint(value),
	}
	return
}

func parseCInstruction(line string) (CInstruction, error) {
	var dest string
	remaining := line
	if before, after, ok := strings.Cut(remaining, "="); ok {
		dest = before
		remaining = after
	}
	comp, jump, _ := strings.Cut(remaining, ";")
	instruction := CInstruction{
		Dest: dest,
		Comp: comp,
		Jump: jump,
	}
	return instruction, nil
}

func parseLInstruction(line string) (instruction LInstruction, err error) {
	remaining, ok1 := strings.CutPrefix(line, "(")
	symbol, ok2 := strings.CutSuffix(remaining, ")")
	if !(ok1 && ok2) {
		err = fmt.Errorf("invalid label declaration: '%s'", string(line))
		return
	}
	instruction = LInstruction{
		Symbol: symbol,
	}
	return
}
