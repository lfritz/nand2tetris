package internal

import (
	"strings"
	"testing"
)

func TestVMWriter(t *testing.T) {
	var output strings.Builder
	w := NewVMWriter(&output)
	w.WritePush(SegmentArgument, 4)
	w.WritePop(SegmentPointer, 5)
	w.WriteArithmetic(CommandSub)
	w.WriteLabel("FOO")
	w.WriteGoto("BAR")
	w.WriteIf("BAZ")
	w.WriteCall("qux", 6)
	w.WriteFunction("Cls.func", 7)
	w.WriteReturn()
	want := `
	push argument 4
	pop pointer 5
	sub
label FOO
	goto BAR
	if-goto BAZ
	call qux 6
function Cls.func 7
	return
`
	want = strings.TrimPrefix(want, "\n")
	got := output.String()
	if got != want {
		t.Errorf(`VMWriter produced:
%s
Want:
%s
`, got, want)
	}
}
