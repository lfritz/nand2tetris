package internal

import (
	"fmt"
	"io"
)

// A VMWriter is used to write Hack VM code.
type VMWriter struct {
	io.Writer
}

// NewVMWriter returns a VMWriter that will write to 'w'.
func NewVMWriter(w io.Writer) *VMWriter {
	return &VMWriter{w}
}

func (w *VMWriter) WritePush(segment Segment, index int) {
	fmt.Fprintf(w, "\tpush %s %d\n", segment, index)
}

func (w *VMWriter) WritePop(segment Segment, index int) {
	fmt.Fprintf(w, "\tpop %s %d\n", segment, index)
}

func (w *VMWriter) WriteArithmetic(c Command) {
	fmt.Fprintf(w, "\t%s\n", c)
}

func (w *VMWriter) WriteLabel(label string) {
	fmt.Fprintf(w, "label %s\n", label)
}

func (w *VMWriter) WriteGoto(label string) {
	fmt.Fprintf(w, "\tgoto %s\n", label)
}

func (w *VMWriter) WriteIf(label string) {
	fmt.Fprintf(w, "\tif-goto %s\n", label)
}

func (w *VMWriter) WriteCall(name string, nArgs int) {
	fmt.Fprintf(w, "\tcall %s %d\n", name, nArgs)
}

func (w *VMWriter) WriteFunction(name string, nVars int) {
	fmt.Fprintf(w, "function %s %d\n", name, nVars)
}

func (w *VMWriter) WriteReturn() {
	fmt.Fprintln(w, "\treturn")
}
