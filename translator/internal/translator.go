package internal

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Run runs the translator. It reads and parses Hack VM instructions from r, translates them to Hack
// assembly code, and writes the result to w.
func Run(filename string, r io.Reader, w io.Writer) error {
	t := NewTranslator(NewInstructionWriter(w, filename), filename)
	parser := NewParser(r)
	for parser.Parse() {
		err := t.translate(parser.Command())
		if err != nil {
			return err
		}
	}
	if err := parser.Err(); err != nil {
		return err
	}
	t.infiniteLoop()
	return nil
}

type Translator struct {
	*InstructionWriter
	filename        string
	currentFunction string
}

func NewTranslator(iw *InstructionWriter, filename string) *Translator {
	return &Translator{iw, filename, ""}
}

func (t *Translator) translate(c Command) error {
	switch c.Type {
	case PushCommand, PopCommand:
		segment := c.Arg1
		index, err := strconv.Atoi(c.Arg2)
		if err != nil {
			return fmt.Errorf("expected decimal number: %s", c.Arg2)
		}
		if c.Type == PushCommand {
			return t.translatePush(segment, index)
		} else {
			return t.translatePop(segment, index)
		}
	case ArithmeticCommand:
		op := c.Arg1
		switch op {
		case "neg", "not":
			t.translateUnaryOperator(op)
			return nil
		case "add", "sub", "eq", "gt", "lt", "and", "or":
			t.translateBinaryOperator(op)
			return nil
		default:
			return fmt.Errorf("unexpected arithmetic-logical command: %q", op)
		}
	case LabelCommand:
		t.translateLabel(c.Arg1)
	case GotoCommand:
		t.translateGoto(c.Arg1)
	case IfCommand:
		t.translateIf(c.Arg1)
	case FunctionCommand:
		nVars, err := strconv.Atoi(c.Arg2)
		if err != nil {
			return fmt.Errorf("expected decimal number: %s", c.Arg2)
		}
		t.translateFunction(c.Arg1, nVars)
	case ReturnCommand:
		t.translateReturn()
	default:
		return fmt.Errorf("unexpected command type: %v", c.Type)
	}
	return nil
}

func (t *Translator) translatePush(segment string, index int) error {
	t.WriteComment("push %s %d", segment, index)
	switch segment {
	case "constant":
		t.WriteADecimal(index)
		t.WriteC("D=A")
	case "static":
		t.WriteASymbolic(fmt.Sprintf("%s.%d", t.filename, index))
		t.WriteC("D=M")
	case "temp", "pointer":
		t.WriteADecimal(segmentAddresses[segment] + index)
		t.WriteC("D=M")
	case "local", "argument", "this", "that":
		base := segmentNames[segment]
		t.WriteASymbolic(base)
		t.WriteC("D=M")
		t.WriteADecimal(index)
		t.WriteC("A=D+A")
		t.WriteC("D=M")
	default:
		return fmt.Errorf("invalid segment name: %q", segment)
	}
	t.push()
	return nil
}

func (t *Translator) translatePop(segment string, index int) error {
	t.WriteComment("pop %s %d", segment, index)
	switch segment {
	case "local", "argument", "this", "that":
		base := segmentNames[segment]
		t.WriteASymbolic(base)
		t.WriteC("D=M")
		t.WriteADecimal(index)
		t.WriteC("D=D+A")
		t.WriteASymbolic("R13")
		t.WriteC("M=D")
	}
	t.pop()
	switch segment {
	case "static":
		t.WriteASymbolic(fmt.Sprintf("%s.%d", index))
		t.WriteC("M=D")
	case "temp", "pointer":
		t.WriteADecimal(segmentAddresses[segment] + index)
		t.WriteC("M=D")
	case "local", "argument", "this", "that":
		t.WriteASymbolic("R13")
		t.WriteC("A=M")
		t.WriteC("M=D")
	}
	return nil
}

var segmentNames = map[string]string{
	"local":    "LCL",
	"argument": "ARG",
	"this":     "THIS",
	"that":     "THAT",
}

var segmentAddresses = map[string]int{
	"pointer": 3,
	"temp":    5,
}

func (t *Translator) translateUnaryOperator(op string) {
	t.WriteComment("%s", op)
	t.pop()
	switch op {
	case "neg":
		t.WriteC("D=-D")
	case "not":
		t.WriteC("D=!D")
	}
	t.push()
}

func (t *Translator) translateBinaryOperator(op string) {
	t.WriteComment("%s", op)
	t.pop()
	t.WriteASymbolic("R13")
	t.WriteC("M=D")
	t.pop()
	t.WriteASymbolic("R13")
	switch op {
	case "add":
		t.WriteC("D=D+M")
	case "sub":
		t.WriteC("D=D-M")
	case "and":
		t.WriteC("D=D&M")
	case "or":
		t.WriteC("D=D|M")
	default:
		l1 := t.NewLabel()
		l2 := t.NewLabel()
		t.WriteC("D=D-M")
		t.WriteASymbolic(l1)
		switch op {
		case "eq":
			t.WriteC("D;JEQ")
		case "gt":
			t.WriteC("D;JGT")
		case "lt":
			t.WriteC("D;JLT")
		}
		t.WriteC("D=0")
		t.WriteASymbolic(l2)
		t.WriteC("0;JMP")
		t.WriteLabel(l1)
		t.WriteC("D=-1")
		t.WriteLabel(l2)
	}
	t.push()
}

// push writes code to push the value in register D on the stack.
func (t *Translator) push() {
	t.WriteASymbolic("SP")
	t.WriteC("A=M")
	t.WriteC("M=D")
	t.WriteASymbolic("SP")
	t.WriteC("M=M+1")
}

// pop writes code to pop the top element of the stack and stores it in register D.
func (t *Translator) pop() {
	t.WriteASymbolic("SP")
	t.WriteC("M=M-1")
	t.WriteC("A=M")
	t.WriteC("D=M")
}

func (t *Translator) translateLabel(label string) {
	t.WriteComment("label %s", label)
	t.WriteLabel(t.buildLabel(label))
}

func (t *Translator) translateGoto(label string) {
	t.WriteComment("goto %s", label)
	t.WriteASymbolic(t.buildLabel(label))
	t.WriteC("0;JMP")
}

func (t *Translator) translateIf(label string) {
	t.WriteComment("if-goto %s", label)
	t.pop()
	t.WriteASymbolic(t.buildLabel(label))
	t.WriteC("D;JNE")
}

func (t *Translator) translateFunction(functionName string, nVars int) {
	t.WriteComment("function %s %d", functionName, nVars)
	t.WriteLabel(functionName)
	for i := 0; i < nVars; i++ {
		t.WriteADecimal(0)
		t.WriteC("D=A")
		t.push()
	}
}

func (t *Translator) translateReturn() {
	t.WriteBlank()
	t.WriteComment("return")

	t.WriteComment("(store LCL in R13)")
	t.WriteASymbolic("LCL")
	t.WriteC("D=M")
	t.WriteASymbolic("R13")
	t.WriteC("M=D")

	t.WriteComment("(store return address in R14)")
	t.WriteADecimal(5)
	t.WriteC("A=D-A")
	t.WriteC("D=M")
	t.WriteASymbolic("R14")
	t.WriteC("M=D")

	t.WriteComment("(*ARG = pop())")
	t.pop()
	t.WriteASymbolic("ARG")
	t.WriteC("A=M")
	t.WriteC("M=D")

	t.WriteComment("(SP = ARG + 1)")
	t.WriteC("D=A+1")
	t.WriteASymbolic("SP")
	t.WriteC("A=M")
	t.WriteC("M=D")

	for _, register := range strings.Split("THAT THIS ARG LCL", " ") {
		t.WriteComment("(decrement R13)")
		t.WriteASymbolic("R13")
		t.WriteC("D=M-1")
		t.WriteC("M=D")
		t.WriteComment("(%s = *R13)", register)
		t.WriteC("A=D")
		t.WriteC("D=M")
		t.WriteASymbolic(register)
		t.WriteC("A=M")
		t.WriteC("M=D")
	}

	t.WriteComment("(goto *R14)")
	t.WriteASymbolic("R14")
	t.WriteC("A=M;JMP")

	t.WriteBlank()
}

func (t *Translator) buildLabel(label string) string {
	return fmt.Sprintf("%s.%s$%s", t.filename, t.currentFunction, label)
}

func (t *Translator) infiniteLoop() {
	t.WriteComment("infinite loop")
	label := t.NewLabel()
	t.WriteASymbolic(label)
	t.WriteLabel(label)
	t.WriteC("0;JMP")
}
