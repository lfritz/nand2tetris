package internal

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func sampleProgram() io.Reader {
	source := `
		(first)
		M=1
		@second
		(second)
		@variable
		@first
		@anothervariable
	`
	return strings.NewReader(source)
}

func TestCreateSymbolTable(t *testing.T) {
	table, err := createSymbolTable(sampleProgram())
	if err != nil {
		t.Fatalf("createSymbolTable returned error: %v", err)
	}
	want := predefinedSymbols()
	want["first"] = 0
	want["second"] = 2
	if !reflect.DeepEqual(table, want) {
		t.Errorf("createSymbolTable returned:\n%#v\nwant:\n%#v", table, want)
	}
}

func TestTranslate(t *testing.T) {
	symbolTable := map[string]uint{
		"first":  123,
		"second": 234,
	}
	var builder strings.Builder
	err := translate(sampleProgram(), &builder, symbolTable)
	if err != nil {
		t.Fatalf("translate returned error: %v", err)
	}
	want := `1110111111001000
0000000011101010
0000000000010000
0000000001111011
0000000000010001
`
	got := builder.String()
	if got != want {
		t.Errorf("translate returned:\n%s\nwant\n%s\n", got, want)
	}
}
