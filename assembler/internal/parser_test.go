package internal

import (
	"testing"
)

func TestParser(t *testing.T) {
	input := `
		// a label
		(start)
		// an A-instruction
		@index
		// an empty line
		
		// a C-instruction
		M=1
	`
	parser := NewParser([]byte(input))

	{
		if !parser.Scan() {
			t.Fatal("parser.Scan() returned false after 0 instructions, want true")
		}
		if parser.InstructionType() != TypeL {
			t.Fatalf("got instruction type %v, want %v", parser.InstructionType(), TypeL)
		}
		got, err := parser.LInstruction()
		if err != nil {
			t.Fatalf("parser.LInstruction() returned error: %v", err)
		}
		want := LInstruction{"start"}
		if got != want {
			t.Errorf("parser.LInstruction() returned %#v, want %#v", got, want)
		}
	}

	{
		if !parser.Scan() {
			t.Fatal("parser.Scan() returned false after 1 instructions, want true")
		}
		if parser.InstructionType() != TypeASymbolic {
			t.Fatalf("got instruction type %v, want %v", parser.InstructionType(), TypeASymbolic)
		}
		got, err := parser.SymbolicAInstruction()
		if err != nil {
			t.Fatalf("parser.SymbolicAInstruction() returned error: %v", err)
		}
		want := SymbolicAInstruction{"index"}
		if got != want {
			t.Errorf("parser.SymbolicAInstruction() returned %#v, want %#v", got, want)
		}
	}

	{
		if !parser.Scan() {
			t.Fatal("parser.Scan() returned false after 2 instructions, want true")
		}
		if parser.InstructionType() != TypeC {
			t.Fatalf("got instruction type %v, want %v", parser.InstructionType(), TypeC)
		}
		got, err := parser.CInstruction()
		if err != nil {
			t.Fatalf("parser.CInstruction() returned error: %v", err)
		}
		want := CInstruction{"M", "1", ""}
		if got != want {
			t.Errorf("parser.CInstruction() returned %#v, want %#v", got, want)
		}
	}
}

func TestValidSymbol(t *testing.T) {
	// valid symbols
	cases := []string{
		"foo",
		"FOO",
		"foo123",
		"_123.$:",
	}
	for _, c := range cases {
		if !validSymbol([]rune(c)) {
			t.Errorf("validSymbol(%q) returned false, want true", c)
		}
	}

	// invalid symbols
	cases = []string{
		"",
		"123",
		"hello!",
	}
	for _, c := range cases {
		if validSymbol([]rune(c)) {
			t.Errorf("validSymbol(%q) returned true, want false", c)
		}
	}
}

func TestParseSymbolicAInstruction(t *testing.T) {
	cases := []struct {
		line string
		want SymbolicAInstruction
	}{
		{"@mylabel", SymbolicAInstruction{"mylabel"}},
		{"@.$1", SymbolicAInstruction{".$1"}},
	}
	for _, c := range cases {
		got, err := parseSymbolicAInstruction([]rune(c.line))
		if err != nil {
			t.Errorf("parseSymbolicAInstruction(%q) returned error: %v", c.line, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseSymbolicAInstruction(%q) returned %q, want %q", c.line, got, c.want)
		}
	}
}

func TestParseCInstruction(t *testing.T) {
	cases := []struct {
		line string
		want CInstruction
	}{
		{"0", CInstruction{"", "0", ""}},
		{"D=A", CInstruction{"D", "A", ""}},
		{"ADM=-1", CInstruction{"ADM", "-1", ""}},
		{"AM=D|A", CInstruction{"AM", "D|A", ""}},
		{"0;JMP", CInstruction{"", "0", "JMP"}},
		{"D;JGT", CInstruction{"", "D", "JGT"}},
		{"-A;JLT", CInstruction{"", "-A", "JLT"}},
		{"AD=D&A;JEQ", CInstruction{"AD", "D&A", "JEQ"}},
	}
	for _, c := range cases {
		got, err := parseCInstruction([]rune(c.line))
		if err != nil {
			t.Errorf("parseCInstruction(%q) returned error: %v", c.line, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseCInstruction(%q) returned %q, want %q", c.line, got, c.want)
		}
	}
}

func TestParseLInstruction(t *testing.T) {
	cases := []struct {
		line string
		want LInstruction
	}{
		{"(foo)", LInstruction{"foo"}},
		{"($123_ABC.def:)", LInstruction{"$123_ABC.def:"}},
	}
	for _, c := range cases {
		got, err := parseLInstruction([]rune(c.line))
		if err != nil {
			t.Errorf("parseLInstruction(%q) returned error: %v", c.line, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseLInstruction(%q) returned %q, want %q", c.line, got, c.want)
		}
	}
}
