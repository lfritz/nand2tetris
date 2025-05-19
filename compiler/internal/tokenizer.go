package internal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type Tokenizer struct {
	reader        *bufio.Reader
	current, next rune

	err error

	tt         TokenType
	keyword    Keyword
	symbol     rune
	identifier string
	intVal     int
	stringVal  string
}

func NewTokenizer(r io.Reader) *Tokenizer {
	t := &Tokenizer{
		reader: bufio.NewReader(r),
	}
	t.advance()
	t.advance()
	return t
}

func (t *Tokenizer) atEnd() bool {
	return t.current == 0
}

func (t *Tokenizer) errorf(format string, a ...any) {
	if t.err != nil {
		return
	}
	t.err = fmt.Errorf(format, a...)
}

func (t *Tokenizer) advance() rune {
	if t.err != nil {
		return 0
	}

	previous := t.current
	t.current = t.next

	r, _, err := t.reader.ReadRune()
	if err == io.EOF {
		t.next = 0
	} else if err != nil {
		t.err = err
		return 0
	}
	if r == unicode.ReplacementChar {
		t.errorf("invalid UTF-8 input")
		return 0
	}
	t.next = r

	return previous
}

func (t *Tokenizer) Tokenize() bool {
	if t.atEnd() || t.err != nil {
		return false
	}
	t.skipWhitespace()
	if t.atEnd() || t.err != nil {
		return false
	}

	switch {
	case isSymbol(t.current):
		return t.tokenizeSymbol()
	case isDigit(t.current):
		return t.tokenizeNumber()
	case t.current == '"':
		return t.tokenizeString()
	case isIdentifierHead(t.current):
		return t.tokenizeIdentifierOrKeyword()
	}
	t.errorf("unexpected character: '%c'", t.current)
	return false
}

func (t *Tokenizer) tokenizeSymbol() bool {
	t.tt = TokenTypeSymbol
	t.symbol = t.current
	t.advance()
	return true
}

func (t *Tokenizer) tokenizeNumber() bool {
	var b strings.Builder
	b.WriteRune(t.advance())
	for isDigit(t.current) {
		b.WriteRune(t.advance())
	}
	n, err := strconv.Atoi(b.String())
	if err != nil {
		t.errorf("invalid integer constant: %q", b.String())
		return false
	}
	t.tt = TokenTypeIntConst
	t.intVal = n
	return true
}

func (t *Tokenizer) tokenizeString() bool {
	t.advance()
	var b strings.Builder
	for !t.atEnd() {
		r := t.advance()
		if r == 0 {
			return false
		}
		if r == '"' {
			t.tt = TokenTypeStringConst
			t.stringVal = b.String()
			return true
		}
		b.WriteRune(r)
	}
	return false
}

func (t *Tokenizer) tokenizeIdentifierOrKeyword() bool {
	var b strings.Builder
	b.WriteRune(t.advance())
	for !t.atEnd() && isIdentifierTail(t.current) {
		b.WriteRune(t.advance())
	}

	text := b.String()
	if kw, ok := keywords[text]; ok {
		t.tt = TokenTypeKeyword
		t.keyword = kw
	} else {
		t.tt = TokenTypeIdentifier
		t.identifier = text
	}
	return true
}

func isSymbol(r rune) bool {
	switch r {
	case '(', ')', '{', '}', '[', ']', '.', ',', ';', '+', '-', '*', '/', '&', '|', '<', '>', '=', '~':
		return true
	}
	return false
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isIdentifierHead(r rune) bool {
	return r == '_' || isLetter(r)
}

func isIdentifierTail(r rune) bool {
	return isIdentifierHead(r) || isDigit(r)
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func (t *Tokenizer) skipWhitespace() {
	for !t.atEnd() {
		if t.current == '/' && t.next == '*' {
			t.skipMultilineComment()
		} else if t.current == '/' && t.next == '/' {
			t.skipComment()
		} else {
			switch t.current {
			case ' ', '\t', '\r', '\n':
			default:
				return
			}
			t.advance()
		}
	}
}

func (t *Tokenizer) skipMultilineComment() {
	t.advance()
	t.advance()
	for !t.atEnd() {
		if t.current == '*' && t.next == '/' {
			t.advance()
			t.advance()
			return
		}
		t.advance()
	}
	t.err = errors.New("unterminated comment")
}

func (t *Tokenizer) skipComment() {
	t.advance()
	t.advance()
	for !t.atEnd() && t.current != '\n' {
		t.advance()
	}
}

func (t *Tokenizer) Err() error {
	return t.err
}

func (t *Tokenizer) TokenType() TokenType {
	return t.tt
}

func (t *Tokenizer) Keyword() Keyword {
	return t.keyword
}

func (t *Tokenizer) Symbol() rune {
	return t.symbol
}

func (t *Tokenizer) Identifier() string {
	return t.identifier
}

func (t *Tokenizer) IntVal() int {
	return t.intVal
}

func (t *Tokenizer) StringVal() string {
	return t.stringVal
}
