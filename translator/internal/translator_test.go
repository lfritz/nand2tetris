package internal

import (
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	vmCode := `
// Add two constants.
push constant 2
push constant 3
add
`
	want := `` // TODO write assembly code
	reader := strings.NewReader(vmCode)
	var builder strings.Builder
	err := Run(reader, &builder)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	got := builder.String()
	if got != want {
		t.Errorf("Run produced:\n%s\nWant:\n%s\n", got, want)
	}
}
