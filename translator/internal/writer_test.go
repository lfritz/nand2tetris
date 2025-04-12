package internal

import (
	"strings"
	"testing"
)

func TestInstructionWriter(t *testing.T) {
	var output strings.Builder
	w := NewInstructionWriter(&output)
	w.WriteComment("some instructions")
	w.WriteADecimal(123)
	w.WriteC("A=M;JMP")
	w.WriteASymbolic("foo")
	w.WriteC("AM=M-1")
	w.WriteC("D;JLE")
	w.WriteC("0")

	want := `// some instructions
@123
A=M;JMP
@foo
AM=M-1
D;JLE
0
`
	got := output.String()
	if got != want {
		t.Errorf("InstructionWriter produced:\n%s\nWant:\n%s\n", got, want)
	}
}
