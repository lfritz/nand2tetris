package internal

import "io"

// Run runs the assembler. It reads and parses Hack assembly from r, translates it to Hack binary
// code, and writes the result to w.
func Run(r io.ReadSeeker, w io.Writer) error {
	// The assembler is a two-pass assembler:

	// The first pass creates a symbol table.
	symbolTable, err := createSymbolTable(r)
	if err != nil {
		return err
	}

	// The second pass translate assembly to binary code.
	_, err = r.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	err = translate(r, w, symbolTable)
	if err != nil {
		return err
	}

	return nil
}

func predefinedSymbols() map[string]uint {
	return map[string]uint{
		"R0":     0,
		"R1":     1,
		"R2":     2,
		"R3":     3,
		"R4":     4,
		"R5":     5,
		"R6":     6,
		"R7":     7,
		"R8":     8,
		"R9":     9,
		"R10":    10,
		"R11":    11,
		"R12":    12,
		"R13":    13,
		"R14":    14,
		"R15":    15,
		"SP":     0,
		"LCL":    1,
		"ARG":    2,
		"THIS":   3,
		"THAT":   4,
		"SCREEN": 16384,
		"KBD":    24576,
	}
}

func createSymbolTable(r io.Reader) (map[string]uint, error) {
	symbolTable := predefinedSymbols()
	var address uint
	p := NewParser(r)
	for p.Scan() {
		switch p.InstructionType() {
		case TypeADecimal, TypeASymbolic, TypeC:
			address++
		case TypeL:
			instruction, err := p.LInstruction()
			if err != nil {
				return nil, err
			}
			symbolTable[instruction.Symbol] = address
		}
	}
	return symbolTable, nil
}

func translate(r io.Reader, w io.Writer, symbolTable map[string]uint) error {
	p := NewParser(r)
	hackWriter := NewHackWriter(w)

	var nextAddress uint = 16
	for p.Scan() {
		switch p.InstructionType() {
		case TypeASymbolic:
			instruction, err := p.SymbolicAInstruction()
			if err != nil {
				return err
			}
			value, ok := symbolTable[instruction.Symbol]
			if !ok {
				value = nextAddress
				symbolTable[instruction.Symbol] = value
				nextAddress++
			}
			err = hackWriter.AInstruction(DecimalAInstruction{
				Value: value,
			})
			if err != nil {
				return err
			}
		case TypeADecimal:
			instruction, err := p.DecimalAInstruction()
			if err != nil {
				return err
			}
			err = hackWriter.AInstruction(instruction)
			if err != nil {
				return err
			}
		case TypeC:
			instruction, err := p.CInstruction()
			if err != nil {
				return err
			}
			err = hackWriter.CInstruction(instruction)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
