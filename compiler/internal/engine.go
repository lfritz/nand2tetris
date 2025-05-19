package internal

import (
	"fmt"
	"io"
)

type CompilationEngine struct {
	tokenizer *Tokenizer
	w         io.Writer
	xmlMode   bool

	atEnd bool
	err   error
}

func NewCompilationEngine(r io.Reader, w io.Writer, xmlMode bool) *CompilationEngine {
	tokenizer := NewTokenizer(r)
	return &CompilationEngine{
		tokenizer: tokenizer,
		w:         w,
		xmlMode:   xmlMode,
	}
}

func (e *CompilationEngine) Err() error {
	return e.err
}

func (e *CompilationEngine) tokenType() TokenType {
	return e.tokenizer.TokenType()
}
func (e *CompilationEngine) keyword() Keyword {
	return e.tokenizer.Keyword()
}
func (e *CompilationEngine) symbol() rune {
	return e.tokenizer.Symbol()
}
func (e *CompilationEngine) identifier() string {
	return e.tokenizer.Identifier()
}
func (e *CompilationEngine) intVal() int {
	return e.tokenizer.IntVal()
}
func (e *CompilationEngine) stringVal() string {
	return e.tokenizer.StringVal()
}

func (e *CompilationEngine) setError(err error) {
	if e.err != nil {
		return
	}
	e.err = err
}

func (e *CompilationEngine) errorf(format string, a ...any) {
	e.setError(fmt.Errorf(format, a...))
}

func (e *CompilationEngine) advance() {
	ok := e.tokenizer.Tokenize()
	if !ok {
		if err := e.tokenizer.Err(); err != nil {
			e.setError(err)
			return
		}
		e.atEnd = true
		return
	}
}

func (e *CompilationEngine) expect(t TokenType) {
	if e.atEnd {
		e.errorf("expected %s, reached end of input", t)
		return
	}
	if e.tokenType() != t {
		e.errorf("expected %s, encountered %s", t, e.tokenType())
		return
	}
}

func (e *CompilationEngine) expectKeyword(expected ...Keyword) {
	if e.atEnd {
		e.errorf("expected keyword “%s”, reached end of input", k)
		return
	}
	if e.tokenType() != TokenTypeKeyword {
		e.errorf("expected keyword “%s”, encountered %s", k, e.tokenType())
		return
	}
	if e.keyword() != k {
		e.errorf("expected keyword “%s”, encountered keyword “%s”", k, e.keyword())
		return
	}
}

func (e *CompilationEngine) expectSymbol(s rune) {
	if e.atEnd {
		e.errorf("expected symbol “%c”, reached end of input", s)
		return
	}
	if e.tokenType() != TokenTypeSymbol {
		e.errorf("expected symbol “%c”, encountered %s", s, e.tokenType())
		return
	}
	if e.symbol() != s {
		e.errorf("expected symbol “%c”, encountered keyword “%s”", s, e.symbol())
		return
	}
}

func (e *CompilationEngine) CompileFile() {
	e.advance()
	e.CompileClass()
}

func (e *CompilationEngine) CompileClass() {
	if e.err != nil {
		return
	}
	e.expectKeyword(KeywordClass)
	e.advance()
	e.expect(TokenTypeIdentifier)
	e.advance()
	e.expectSymbol('{')
	for {
		if e.err != nil {
			return
		}
		e.advance()
		if e.tokenType() == TokenTypeSymbol && e.symbol() == '}' {
			break
		}
		e.CompileClassVarDec()
		e.CompileSubroutineDec()
	}
	return
}

func (e *CompilationEngine) CompileClassVarDec() {
	if e.err != nil {
		return
	}
	e.expectKeyword(KeywordStatic, KeywordField)
	var isStatic bool
	switch e.Keyword {
	case KeywordStatic:
		isStatic = true
	case KeywordField:
	default:
		e.errorf("expected “static” or “field”")
		return
	}
	e.CompileType()
	e.CompileVarName()
	for {
		if e.err != nil {
			return
		}
		e.advance()
		if !(e.tokenType() == TokenTypeSymbol && e.symbol() == ',') {
			break
		}
		e.CompileVarName()
	}
	e.expectSymbol(';')
	e.advance()
}

func (e *CompilationEngine) CompileType() {
	if e.err != nil {
		return
	}
	// TODO
}

func (e *CompilationEngine) CompileSubroutineDec() {
	if e.err != nil {
		return
	}
	// TODO
}

func (e *CompilationEngine) CompileVarName() {
	if e.err != nil {
		return
	}
	// TODO
}
