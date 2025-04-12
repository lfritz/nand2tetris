package internal

import (
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	vmCode := `
// Add two constants.
push constant 2
push constant 3
add
`
	want := `// push constant 2
@2
D=A
@SP
A=M
M=D
@SP
M=M+1
// push constant 3
@3
D=A
@SP
A=M
M=D
@SP
M=M+1
// add
@SP
M=M-1
A=M
D=M
@R13
M=D
@SP
M=M-1
A=M
D=M
@R13
D=D+M
@SP
A=M
M=D
@SP
M=M+1
// infinite loop
@l1
(l1)
0;JMP
`
	reader := strings.NewReader(vmCode)
	var builder strings.Builder
	err := Run("filename", reader, &builder)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	got := builder.String()
	if got != want {
		t.Errorf("Run produced:\n%s\nWant:\n%s\n", got, want)
	}
}

func TestTranslate(t *testing.T) {
	cases := []struct {
		command Command
		want    string
	}{
		{Command{PushCommand, "constant", "11"}, `
			// push constant 11
			@11
			D=A
			@SP
			A=M
			M=D
			@SP
			M=M+1`},
		{Command{PushCommand, "static", "13"}, `
			// push static 13
			@filename.13
			D=M
			@SP
			A=M
			M=D
			@SP
			M=M+1`},
		{Command{PushCommand, "temp", "7"}, `
			// push temp 7
			@12
			D=M
			@SP
			A=M
			M=D
			@SP
			M=M+1`},
		{Command{PushCommand, "local", "17"}, `
			// push local 17
			@LCL
			D=M
			@17
			A=D+A
			D=M
			@SP
			A=M
			M=D
			@SP
			M=M+1`},
		{Command{PopCommand, "static", "13"}, `
			// pop static 13
			@SP
			M=M-1
			A=M
			D=M
			@filename.13
			M=D`},
		{Command{PopCommand, "pointer", "1"}, `
			// pop pointer 1
			@SP
			M=M-1
			A=M
			D=M
			@4
			M=D`},
		{Command{PopCommand, "argument", "3"}, `
			// pop argument 3
			@ARG
			D=M
			@3
			D=D+A
			@R13
			M=D
			@SP
			M=M-1
			A=M
			D=M
			@R13
			A=M
			M=D`},
		{Command{ArithmeticCommand, "neg", ""}, `
			// neg
			@SP
			M=M-1
			A=M
			D=M
			D=-D
			@SP
			A=M
			M=D
			@SP
			M=M+1`},
		{Command{ArithmeticCommand, "sub", ""}, `
			// sub
			@SP
			M=M-1
			A=M
			D=M
			@R13
			M=D
			@SP
			M=M-1
			A=M
			D=M
			@R13
			D=D-M
			@SP
			A=M
			M=D
			@SP
			M=M+1`},
		{Command{ArithmeticCommand, "eq", ""}, `
			// eq
			@SP
			M=M-1
			A=M
			D=M
			@R13
			M=D
			@SP
			M=M-1
			A=M
			D=M
			@R13
			D=D-M
			@l1
			D;JEQ
			D=0
			@l2
			0;JMP
			(l1)
			D=-1
			(l2)
			@SP
			A=M
			M=D
			@SP
			M=M+1`},
	}

	for _, c := range cases {
		want := c.want
		want = strings.TrimLeft(want, "\n")
		want = strings.ReplaceAll(want, "\t", "")
		want = want + "\n"
		var output strings.Builder
		err := translate("filename", c.command, NewInstructionWriter(&output))
		if err != nil {
			t.Errorf("translate for\n%#v\nreturned error: %v", c.command, err)
			continue
		}
		got := output.String()
		if got != want {
			t.Errorf("translate for\n%#v\nproduced:\n%s\nWant:\n%s\n", c.command, got, want)
		}
	}
}
