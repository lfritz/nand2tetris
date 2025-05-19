package internal

import (
	"fmt"
	"io"
)

// Compile runs the compiler and writes Hack VM code to w.
func Compile(filename string, r io.Reader, w io.Writer) error {
	engine := NewCompilationEngine(r, w, false)
	engine.CompileClass()
	return engine.Err()
}

// PrintSyntax runs the tokenizer and writes tokens to w.
func PrintTokens(filename string, r io.Reader, w io.Writer) error {
	t := NewTokenizer(r)
	fmt.Fprintf(w, "<tokens>\n")
	for t.Tokenize() {
		switch t.TokenType() {
		case TokenTypeKeyword:
			fmt.Fprintf(w, "<keyword> %s </keyword>\n", t.Keyword().String())
		case TokenTypeSymbol:
			fmt.Fprintf(w, "<symbol> %s </symbol>\n", symbolToXML(t.Symbol()))
		case TokenTypeIdentifier:
			fmt.Fprintf(w, "<identifier> %s </identifier>\n", t.Identifier())
		case TokenTypeIntConst:
			fmt.Fprintf(w, "<integerConstant> %d </integerConstant>\n", t.IntVal())
		case TokenTypeStringConst:
			fmt.Fprintf(w, "<stringConstant> %s </stringConstant>\n", t.StringVal())
		}
	}
	if err := t.Err(); err != nil {
		return err
	}
	fmt.Fprintf(w, "</tokens>\n")
	return nil
}

// PrintSyntax runs the parser and writes a syntax tree to w.
func PrintSyntax(filename string, r io.Reader, w io.Writer) error {
	engine := NewCompilationEngine(r, w, true)
	engine.CompileClass()
	return engine.Err()
}

func symbolToXML(r rune) string {
	switch r {
	case '<':
		return "&lt;"
	case '>':
		return "&gt;"
	case '"':
		return "&quot;"
	case '&':
		return "&amp;"
	}
	return fmt.Sprintf("%c", r)
}
