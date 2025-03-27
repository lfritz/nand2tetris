package internal

// InstructionType is an enum for the different types of instructions in Hack assembly programs.
type InstructionType int

const (
	TypeADecimal InstructionType = iota
	TypeASymbolic
	TypeC
	TypeL
)

// A SymbolicAInstruction is an A-instruction that contains a symbol, for example "@START".
type SymbolicAInstruction struct {
	Symbol string
}

// A DecimalAInstruction is an A-instructions that contains a concrete number, for example "@20".
type DecimalAInstruction struct {
	Value uint
}

// A CInstruction represents a C-instruction with dest, comp, and jump parts, for example
// "A=D;JGT". Dest and Jump may be empty strings.
type CInstruction struct {
	Dest, Comp, Jump string
}

// An LInstruction represents a label pseudo-instruction, for example "(START)".
type LInstruction struct {
	Symbol string
}
