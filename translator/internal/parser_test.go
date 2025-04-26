package internal

import (
	"reflect"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	input := `
add  // an arithmetic command

// push and pop commands
push temp 0
pop static 8
`
	parser := NewParser(strings.NewReader(input))

	if !parser.Parse() {
		t.Fatalf("parser.Parse() returned false after 0 instructions\nerror: %v", parser.Err())
	}
	got := parser.Command()
	want := Command{Type: ArithmeticCommand, Arg1: "add"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parser returned %v, want %v", got, want)
	}

	if !parser.Parse() {
		t.Fatalf("parser.Parse() returned false after 1 instructions\nerror: %v", parser.Err())
	}
	got = parser.Command()
	want = Command{Type: PushCommand, Arg1: "temp", Arg2: "0"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parser returned %v, want %v", got, want)
	}

	if !parser.Parse() {
		t.Fatalf("parser.Parse() returned false after 2 instructions\nerror: %v", parser.Err())
	}
	got = parser.Command()
	want = Command{Type: PopCommand, Arg1: "static", Arg2: "8"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parser returned %v, want %v", got, want)
	}

	if parser.Parse() {
		t.Error("parser.Parse() returned true after 3 instructions, want false")
	}
	if err := parser.Err(); err != nil {
		t.Errorf("parser.Err() returned %v", err)
	}
}

func TestParseCommand(t *testing.T) {
	cases := []struct {
		line string
		want *Command
	}{
		{"add", &Command{Type: ArithmeticCommand, Arg1: "add"}},
		{"push local 2", &Command{Type: PushCommand, Arg1: "local", Arg2: "2"}},
		{"pop this 510", &Command{Type: PopCommand, Arg1: "this", Arg2: "510"}},
	}
	for _, c := range cases {
		got, err := parseCommand(c.line)
		if err != nil {
			t.Errorf("parseCommand(%q) returned error: %v", c.line, err)
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("parseCommand(%q) returned %v, want %v", c.line, got, c.want)
		}
	}
}

func TestParseCommandInvalid(t *testing.T) {
	cases := []string{
		"hello",
		"add 1",
		"push 2",
	}
	for _, c := range cases {
		_, err := parseCommand(c)
		if err == nil {
			t.Errorf("parseCommand(%q) did not return error", c)
		}
	}
}
