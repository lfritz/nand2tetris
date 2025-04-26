package internal

import "fmt"

// CommandType is an enum for the different types of command in the Hack VM language.
type CommandType int

const (
	InvalidCommand CommandType = iota
	ArithmeticCommand
	PushCommand
	PopCommand
	LabelCommand
	GotoCommand
	IfCommand
	FunctionCommand
	ReturnCommand
	CallCommand
)

func (c CommandType) String() string {
	switch c {
	case InvalidCommand:
		return "<invalid>"
	case ArithmeticCommand:
		return "arithmetic"
	case PushCommand:
		return "push"
	case PopCommand:
		return "pop"
	case LabelCommand:
		return "label"
	case GotoCommand:
		return "goto"
	case IfCommand:
		return "if-goto"
	case FunctionCommand:
		return "function"
	case ReturnCommand:
		return "return"
	case CallCommand:
		return "call"
	default:
		return fmt.Sprintf("<command type %d>", c)
	}
}

// arity returns the number of arguments expected by a command.
func (c CommandType) arity() int {
	switch c {
	case ArithmeticCommand:
		return 0
	case PushCommand, PopCommand:
		return 2
	case LabelCommand:
		return 1
	case GotoCommand:
		return 1
	case IfCommand:
		return 1
	}
	return 0
}

func commandType(name string) CommandType {
	switch name {
	case "add", "sub", "neg", "eq", "gt", "lt", "and", "or", "not":
		return ArithmeticCommand
	case "push":
		return PushCommand
	case "pop":
		return PopCommand
	case "label":
		return LabelCommand
	case "goto":
		return GotoCommand
	case "if-goto":
		return IfCommand
	}
	return InvalidCommand
}

// A Command represents a command in the Hack VM language.
//
// For arithmetic-logical commands, Arg1 contains the actual command and Arg2 is empty.
//
// For a 'label', 'goto', or 'if-goto' command, Arg1 contains the label and Arg2 is empty.
type Command struct {
	Type       CommandType
	Arg1, Arg2 string
}
