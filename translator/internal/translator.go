package internal

import (
	"fmt"
	"io"
	"strconv"
)

// Run runs the translator. It reads and parses Hack VM instructions from r, translates them to Hack
// assembly code, and writes the result to w.
func Run(filename string, r io.Reader, w io.Writer) error {
	iw := NewInstructionWriter(w, filename)
	currentFunction := "" // TODO
	parser := NewParser(r)
	for parser.Parse() {
		err := translate(filename, currentFunction, parser.Command(), iw)
		if err != nil {
			return err
		}
	}
	if err := parser.Err(); err != nil {
		return err
	}
	infiniteLoop(iw)
	return nil
}

func translate(filename, currentFunction string, c Command, w *InstructionWriter) error {
	switch c.Type {
	case PushCommand, PopCommand:
		segment := c.Arg1
		index, err := strconv.Atoi(c.Arg2)
		if err != nil {
			return fmt.Errorf("expected decimal number: %s", c.Arg2)
		}
		if c.Type == PushCommand {
			return translatePush(filename, segment, index, w)
		} else {
			return translatePop(filename, segment, index, w)
		}
	case ArithmeticCommand:
		op := c.Arg1
		switch op {
		case "neg", "not":
			translateUnaryOperator(op, w)
			return nil
		case "add", "sub", "eq", "gt", "lt", "and", "or":
			translateBinaryOperator(op, w)
			return nil
		default:
			return fmt.Errorf("unexpected arithmetic-logical command: %q", op)
		}
	case LabelCommand:
		translateLabel(filename, currentFunction, c.Arg1, w)
		return nil
	case GotoCommand:
		translateGoto(filename, currentFunction, c.Arg1, w)
		return nil
	case IfCommand:
		translateIf(filename, currentFunction, c.Arg1, w)
		return nil
	}
	return fmt.Errorf("unexpected command type: %v", c.Type)
}

func translatePush(filename, segment string, index int, w *InstructionWriter) error {
	w.WriteComment("push %s %d", segment, index)
	switch segment {
	case "constant":
		w.WriteADecimal(index)
		w.WriteC("D=A")
	case "static":
		w.WriteASymbolic(fmt.Sprintf("%s.%d", filename, index))
		w.WriteC("D=M")
	case "temp", "pointer":
		w.WriteADecimal(segmentAddresses[segment] + index)
		w.WriteC("D=M")
	case "local", "argument", "this", "that":
		base := segmentNames[segment]
		w.WriteASymbolic(base)
		w.WriteC("D=M")
		w.WriteADecimal(index)
		w.WriteC("A=D+A")
		w.WriteC("D=M")
	default:
		return fmt.Errorf("invalid segment name: %q", segment)
	}
	push(w)
	return nil
}

func translatePop(filename, segment string, index int, w *InstructionWriter) error {
	w.WriteComment("pop %s %d", segment, index)
	switch segment {
	case "local", "argument", "this", "that":
		base := segmentNames[segment]
		w.WriteASymbolic(base)
		w.WriteC("D=M")
		w.WriteADecimal(index)
		w.WriteC("D=D+A")
		w.WriteASymbolic("R13")
		w.WriteC("M=D")
	}
	pop(w)
	switch segment {
	case "static":
		w.WriteASymbolic(fmt.Sprintf("%s.%d", filename, index))
		w.WriteC("M=D")
	case "temp", "pointer":
		w.WriteADecimal(segmentAddresses[segment] + index)
		w.WriteC("M=D")
	case "local", "argument", "this", "that":
		w.WriteASymbolic("R13")
		w.WriteC("A=M")
		w.WriteC("M=D")
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

func translateUnaryOperator(op string, w *InstructionWriter) {
	w.WriteComment("%s", op)
	pop(w)
	switch op {
	case "neg":
		w.WriteC("D=-D")
	case "not":
		w.WriteC("D=!D")
	}
	push(w)
}

func translateBinaryOperator(op string, w *InstructionWriter) {
	w.WriteComment("%s", op)
	pop(w)
	w.WriteASymbolic("R13")
	w.WriteC("M=D")
	pop(w)
	w.WriteASymbolic("R13")
	switch op {
	case "add":
		w.WriteC("D=D+M")
	case "sub":
		w.WriteC("D=D-M")
	case "and":
		w.WriteC("D=D&M")
	case "or":
		w.WriteC("D=D|M")
	default:
		l1 := w.NewLabel()
		l2 := w.NewLabel()
		w.WriteC("D=D-M")
		w.WriteASymbolic(l1)
		switch op {
		case "eq":
			w.WriteC("D;JEQ")
		case "gt":
			w.WriteC("D;JGT")
		case "lt":
			w.WriteC("D;JLT")
		}
		w.WriteC("D=0")
		w.WriteASymbolic(l2)
		w.WriteC("0;JMP")
		w.WriteLabel(l1)
		w.WriteC("D=-1")
		w.WriteLabel(l2)
	}
	push(w)
}

func push(w *InstructionWriter) {
	w.WriteASymbolic("SP")
	w.WriteC("A=M")
	w.WriteC("M=D")
	w.WriteASymbolic("SP")
	w.WriteC("M=M+1")
}

func pop(w *InstructionWriter) {
	w.WriteASymbolic("SP")
	w.WriteC("M=M-1")
	w.WriteC("A=M")
	w.WriteC("D=M")
}

func translateLabel(filename, currentFunction, label string, w *InstructionWriter) {
	w.WriteComment("label %s", label)
	w.WriteLabel(buildLabel(filename, currentFunction, label))
}

func translateGoto(filename, currentFunction, label string, w *InstructionWriter) {
	w.WriteComment("goto %s", label)
	w.WriteASymbolic(buildLabel(filename, currentFunction, label))
	w.WriteC("0;JMP")
}

func translateIf(filename, currentFunction, label string, w *InstructionWriter) {
	w.WriteComment("if-goto %s", label)
	pop(w)
	w.WriteASymbolic(buildLabel(filename, currentFunction, label))
	w.WriteC("D;JNE")
}

func buildLabel(filename, currentFunction, label string) string {
	return fmt.Sprintf("%s.%s$%s", filename, currentFunction, label)
}

func infiniteLoop(w *InstructionWriter) {
	w.WriteComment("infinite loop")
	label := w.NewLabel()
	w.WriteASymbolic(label)
	w.WriteLabel(label)
	w.WriteC("0;JMP")
}
