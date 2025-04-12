package internal

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

// arity returns the number of arguments expected by a command.
func (c CommandType) arity() int {
	switch c {
	case ArithmeticCommand:
		return 0
	case PushCommand, PopCommand:
		return 2
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
	}
	return InvalidCommand
}

// A Command represents a command in the Hack VM language.
//
// For arithmetic-logical commands, Arg1 contains the actual command; Arg2 is empty.
type Command struct {
	Type       CommandType
	Arg1, Arg2 string
}
