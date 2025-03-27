package internal

import (
	"strings"
	"testing"
)

func TestWriteAInstruction(t *testing.T) {
	cases := []struct {
		instruction DecimalAInstruction
		want        string
	}{
		{DecimalAInstruction{0}, "0000000000000000\n"},
		{DecimalAInstruction{1}, "0000000000000001\n"},
		{DecimalAInstruction{12345}, "0011000000111001\n"},
		{DecimalAInstruction{32767}, "0111111111111111\n"},
	}
	for _, c := range cases {
		var builder strings.Builder
		w := NewHackWriter(&builder)
		err := w.AInstruction(c.instruction)
		if err != nil {
			t.Errorf("AInstruction(%#v) returned error: %v", c.instruction, err)
		}
		got := builder.String()
		if got != c.want {
			t.Errorf("AInstruction(%#v) returned %q, want %q", c.instruction, got, c.want)
		}
	}
}

func TestWriteCInstruction(t *testing.T) {
	cases := []struct {
		instruction CInstruction
		want        string
	}{
		{CInstruction{"A", "A-1", ""}, "1110110010100000\n"},
		{CInstruction{"", "D&M", "JGE"}, "1111000000000011\n"},
		{CInstruction{"MD", "1", "JMP"}, "1110111111011111\n"},
		{CInstruction{"DM", "1", "JMP"}, "1110111111011111\n"},
	}
	for _, c := range cases {
		var builder strings.Builder
		w := NewHackWriter(&builder)
		err := w.CInstruction(c.instruction)
		if err != nil {
			t.Errorf("CInstruction(%#v) returned error: %v", c.instruction, err)
		}
		got := builder.String()
		if got != c.want {
			t.Errorf("CInstruction(%#v) returned %q, want %q", c.instruction, got, c.want)
		}
	}
}
