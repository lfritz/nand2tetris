package internal

type SymbolKind int

const (
	SymbolKindStatic SymbolKind = iota
	SymbolKindField
	SymbolKindArg
	SymbolKindVar
	SymbolKinds
)

type symbol struct {
	name  string
	typ   string
	kind  SymbolKind
	index int
}

// A SymbolTable keeps track of the symbols defined in a Jack program.
type SymbolTable struct {
	symbols map[string]symbol
	count   [SymbolKinds]int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		symbols: make(map[string]symbol),
	}
}

func (t *SymbolTable) Reset() {
	t.symbols = make(map[string]symbol)
	for i := range t.count {
		t.count[i] = 0
	}
}

func (t *SymbolTable) Define(name, typ string, kind SymbolKind) {
	index := t.count[kind]
	t.count[kind] += 1
	t.symbols[name] = symbol{
		name:  name,
		typ:   typ,
		kind:  kind,
		index: index,
	}
}

func (t *SymbolTable) VarCount(kind SymbolKind) int {
	return t.count[kind]
}

func (t *SymbolTable) KindOf(name string) SymbolKind {
	return t.symbols[name].kind
}

func (t *SymbolTable) TypeOf(name string) string {
	return t.symbols[name].typ
}

func (t *SymbolTable) IndexOf(name string) int {
	return t.symbols[name].index
}
