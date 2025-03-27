package internal

import (
	"fmt"
	"io"
	"slices"
)

// HackWriter writes binary instructions for Hack programs.
type HackWriter struct {
	w                               io.Writer
	compTable, destTable, jumpTable map[string]string
}

// NewHackWriter returns a HackWriter that writes binary instructions to w.
func NewHackWriter(w io.Writer) *HackWriter {
	return &HackWriter{
		w:         w,
		compTable: compTable(),
		destTable: destTable(),
		jumpTable: jumpTable(),
	}
}

// AInstruction writes an A-instruction.
func (w *HackWriter) AInstruction(i DecimalAInstruction) error {
	_, err := fmt.Fprintf(w.w, "0%015b\n", i.Value)
	if err != nil {
		return err
	}
	return nil
}

// CInstruction writes a C-instruction.
func (w *HackWriter) CInstruction(i CInstruction) error {
	comp, ok := w.compTable[i.Comp]
	if !ok {
		return fmt.Errorf("invalid comp field in C-instruction: %q", i.Comp)
	}
	dest, ok := w.destCode(i.Dest)
	if !ok {
		return fmt.Errorf("invalid dest field in C-instruction: %q", i.Dest)
	}
	jump, ok := w.jumpTable[i.Jump]
	if !ok {
		return fmt.Errorf("invalid jump field in C-instruction: %q", i.Jump)
	}

	_, err := fmt.Fprintf(w.w, "111%s%s%s\n", comp, dest, jump)
	if err != nil {
		return err
	}
	return nil
}

func (w *HackWriter) destCode(value string) (string, bool) {
	// For the 'dest' part of a C-instruction, we need to allow any order. For example,
	// "MD=1" and "DM=1" are equivalent. Sorting the code first means we need only one combination
	// in 'destTable'.
	runes := []rune(value)
	slices.Sort(runes)
	dest, ok := w.destTable[string(runes)]
	return dest, ok
}

func compTable() map[string]string {
	// these are the binary codes for the "a" and "comp" fields in a C-instruction
	return map[string]string{
		// a == 0
		"0":   "0101010",
		"1":   "0111111",
		"-1":  "0111010",
		"D":   "0001100",
		"A":   "0110000",
		"!D":  "0001101",
		"!A":  "0110001",
		"-D":  "0001111",
		"-A":  "0110011",
		"D+1": "0011111",
		"A+1": "0110111",
		"D-1": "0001110",
		"A-1": "0110010",
		"D+A": "0000010",
		"D-A": "0010011",
		"A-D": "0000111",
		"D&A": "0000000",
		"D|A": "0010101",
		// a == 1
		"M":   "1110000",
		"!M":  "1110001",
		"-M":  "1110011",
		"M+1": "1110111",
		"M-1": "1110010",
		"D+M": "1000010",
		"D-M": "1010011",
		"M-D": "1000111",
		"D&M": "1000000",
		"D|M": "1010101",
	}
}

func destTable() map[string]string {
	// these are the binary codes for the "dest" field in a C-instruction
	return map[string]string{
		"":    "000",
		"M":   "001",
		"D":   "010",
		"DM":  "011",
		"A":   "100",
		"AM":  "101",
		"AD":  "110",
		"ADM": "111",
	}
}

func jumpTable() map[string]string {
	// these are the binary codes for the "jump" field in a C-instruction
	return map[string]string{
		"":    "000",
		"JGT": "001",
		"JEQ": "010",
		"JGE": "011",
		"JLT": "100",
		"JNE": "101",
		"JLE": "110",
		"JMP": "111",
	}
}
