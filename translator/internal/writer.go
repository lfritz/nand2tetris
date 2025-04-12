package internal

import (
	"fmt"
	"io"
)

// An InstructionWriter is used to write Hack assembly language instructions.
type InstructionWriter struct {
	io.Writer
	labelSequence int
}

// NewInstructionWriter returns an InstructionWriter that will write to the given writer.
func NewInstructionWriter(w io.Writer) *InstructionWriter {
	return &InstructionWriter{w, 0}
}

// WriteComment writes a comment.
func (w *InstructionWriter) WriteComment(format string, a ...any) {
	fmt.Fprint(w, "// ")
	fmt.Fprintf(w, format, a...)
	fmt.Fprintln(w)
}

// WriteADecimal writes an A-instruction that contains a concrete number, for example "@20".
func (w *InstructionWriter) WriteADecimal(value int) {
	fmt.Fprintf(w, "@%d\n", value)
}

// WriteASymbolic writes an A-instruction that contains a symbol, for example "@START".
func (w *InstructionWriter) WriteASymbolic(value string) {
	fmt.Fprintf(w, "@%s\n", value)
}

// WriteC writes a C-instruction.
func (w *InstructionWriter) WriteC(instruction string) {
	fmt.Fprintln(w, instruction)
}

// WriteLabel writes a label pseudo-instruction.
func (w *InstructionWriter) WriteLabel(label string) {
	fmt.Fprintf(w, "(%s)\n", label)
}

// NewLabel returns a label. Each call returns a different one.
func (w *InstructionWriter) NewLabel() string {
	w.labelSequence++
	return fmt.Sprintf("l%d", w.labelSequence)
}
