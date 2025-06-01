package internal

import "testing"

func fillTable(t *testing.T, table *SymbolTable) {
	t.Helper()

	table.Define("aaa", "string", IdentifierKindField)
	table.Define("bbb", "boolean", IdentifierKindField)
	table.Define("ccc", "string", IdentifierKindField)

	table.Define("ddd", "int", IdentifierKindStatic)

	table.Define("eee", "string", IdentifierKindArg)
	table.Define("fff", "boolean", IdentifierKindArg)

	table.Define("ggg", "string", IdentifierKindVar)
	table.Define("hhh", "string", IdentifierKindVar)
	table.Define("iii", "boolean", IdentifierKindVar)
	table.Define("jjj", "int", IdentifierKindVar)
}

func TestSymbolTableVarCount(t *testing.T) {
	table := NewSymbolTable()
	fillTable(t, table)
	cases := []struct {
		kind IdentifierKind
		want int
	}{
		{IdentifierKindField, 3},
		{IdentifierKindStatic, 1},
		{IdentifierKindArg, 2},
		{IdentifierKindVar, 4},
	}
	for _, c := range cases {
		got := table.VarCount(c.kind)
		if got != c.want {
			t.Errorf("table.VarCount(%q) == %v, want %v", c.kind, got, c.want)
		}
	}
}

func TestSymbolTableKindOf(t *testing.T) {
	table := NewSymbolTable()
	fillTable(t, table)
	name := "eee"
	got := table.KindOf(name)
	want := IdentifierKindArg
	if got != want {
		t.Errorf("table.KindOf(%q) == %v, want %v", name, got, want)
	}
}

func TestSymbolTableTypeOf(t *testing.T) {
	table := NewSymbolTable()
	fillTable(t, table)
	name := "eee"
	got := table.TypeOf(name)
	want := "string"
	if got != want {
		t.Errorf("table.TypeOf(%q) == %q, want %q", name, got, want)
	}
}

func TestSymbolTableIndexOf(t *testing.T) {
	table := NewSymbolTable()
	fillTable(t, table)

	name := "eee"
	got := table.IndexOf(name)
	want := 0
	if got != want {
		t.Errorf("table.IndexOf(%q) == %v, want %v", name, got, want)
	}

	name = "fff"
	got = table.IndexOf(name)
	want = 1
	if got != want {
		t.Errorf("table.IndexOf(%q) == %v, want %v", name, got, want)
	}
}
