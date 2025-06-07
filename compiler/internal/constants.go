package internal

type Segment int

const (
	SegmentConstant Segment = iota
	SegmentArgument
	SegmentLocal
	SegmentStatic
	SegmentThis
	SegmentThat
	SegmentPointer
	SegmentTemp
)

func (s Segment) String() string {
	switch s {
	case SegmentConstant:
		return "constant"
	case SegmentArgument:
		return "argument"
	case SegmentLocal:
		return "local"
	case SegmentStatic:
		return "static"
	case SegmentThis:
		return "this"
	case SegmentThat:
		return "that"
	case SegmentPointer:
		return "pointer"
	case SegmentTemp:
		return "temp"
	}
	return ""
}

type Command int

const (
	CommandAdd Command = iota
	CommandSub
	CommandNeg
	CommandEq
	CommandGt
	CommandLt
	CommandAnd
	CommandOr
	CommandNot
)

func (c Command) String() string {
	switch c {
	case CommandAdd:
		return "add"
	case CommandSub:
		return "sub"
	case CommandNeg:
		return "neg"
	case CommandEq:
		return "eq"
	case CommandGt:
		return "gt"
	case CommandLt:
		return "lt"
	case CommandAnd:
		return "and"
	case CommandOr:
		return "or"
	case CommandNot:
		return "not"
	}
	return ""
}
