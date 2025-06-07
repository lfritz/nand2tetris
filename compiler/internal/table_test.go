package internal

import "testing"

func fillTable(t *testing.T, table *SymbolTable) {
	t.Helper()

	table.Define("aaa", "string", SymbolKindField)
	table.Define("bbb", "boolean", SymbolKindField)
	table.Define("ccc", "string", SymbolKindField)

	table.Define("ddd", "int", SymbolKindStatic)

	table.Define("eee", "string", SymbolKindArg)
	table.Define("fff", "boolean", SymbolKindArg)

	table.Define("ggg", "string", SymbolKindVar)
	table.Define("hhh", "string", SymbolKindVar)
	table.Define("iii", "boolean", SymbolKindVar)
	table.Define("jjj", "int", SymbolKindVar)
}

func TestSymbolTableVarCount(t *testing.T) {
	table := NewSymbolTable()
	fillTable(t, table)
	cases := []struct {
		kind SymbolKind
		want int
	}{
		{SymbolKindField, 3},
		{SymbolKindStatic, 1},
		{SymbolKindArg, 2},
		{SymbolKindVar, 4},
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
	want := SymbolKindArg
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
