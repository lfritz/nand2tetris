package internal

import "io"

// Run runs the translator. It reads and parses Hack VM instructions from r, translates them to Hack
// assembly code, and writes the result to w.
func Run(r io.Reader, w io.Writer) error {
	parser := NewParser(r)
	for parser.Parse() {
		command := parser.Command()
		_ = command
	}
	return parser.Err()
}
