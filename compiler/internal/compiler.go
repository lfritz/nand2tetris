package internal

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// Compile runs the compiler and writes Hack VM code to w.
func Compile(filename string, r io.Reader, w io.Writer) error {
	c := newCompiler(r, w, false)
	return c.compileFile()
}

// PrintTokens runs the tokenizer and writes tokens to w.
func PrintTokens(filename string, r io.Reader, w io.Writer) error {
	t := NewTokenizer(r)
	fmt.Fprintf(w, "<tokens>\n")
	for t.Tokenize() {
		printToken(w, t)
	}
	if err := t.Err(); err != nil {
		return err
	}
	fmt.Fprintf(w, "</tokens>\n")
	return nil
}

func printToken(w io.Writer, t *Tokenizer) {
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

// PrintSyntax runs the parser and writes a syntax tree to w.
func PrintSyntax(filename string, r io.Reader, w io.Writer) error {
	c := newCompiler(r, w, true)
	return c.compileFile()
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

type compiler struct {
	t               *Tokenizer
	w               io.Writer
	printMode       bool
	atEnd           bool
	classTable      *SymbolTable
	subroutineTable *SymbolTable
}

func newCompiler(r io.Reader, w io.Writer, printMode bool) *compiler {
	return &compiler{
		t:               NewTokenizer(r),
		w:               w,
		printMode:       printMode,
		classTable:      NewSymbolTable(),
		subroutineTable: NewSymbolTable(),
	}
}

func (c *compiler) compileFile() error {
	if !c.printMode {
		return errors.New("compilation is not yet implemented")
	}
	c.atEnd = !c.t.Tokenize()
	if err := c.compileClass(); err != nil {
		return err
	}
	if !c.atEnd {
		return errors.New("expected end of file")
	}
	return nil
}

func (c *compiler) compileClass() error {
	fmt.Fprintf(c.w, "<class>\n")
	if err := c.consumeKeyword(KeywordClass); err != nil {
		return err
	}
	className, err := c.consumeIdentifier()
	if err != nil {
		return err
	}
	_ = className
	if err := c.consumeSymbol('{'); err != nil {
		return err
	}
	for !c.gotSymbol('}') {
		if c.gotKeyword(KeywordStatic, KeywordField) {
			err = c.compileClassVarDec()
		} else {
			err = c.compileSubroutine()
		}
		if err != nil {
			return err
		}
	}
	c.advance()
	fmt.Fprintf(c.w, "</class>\n")
	c.classTable.Reset()
	return nil
}

func (c *compiler) compileClassVarDec() error {
	fmt.Fprintf(c.w, "<classVarDec>\n")
	kind := SymbolKindField
	if c.gotKeyword(KeywordStatic) {
		kind = SymbolKindStatic
	}
	if err := c.consumeKeyword(KeywordStatic, KeywordField); err != nil {
		return err
	}
	typ, err := c.compileType()
	if err != nil {
		return err
	}
	for {
		name, err := c.consumeIdentifier()
		if err != nil {
			return err
		}
		c.classTable.Define(name, typ, kind)
		if c.gotSymbol(',') {
			c.advance()
		} else {
			break
		}
	}
	if err := c.consumeSymbol(';'); err != nil {
		return err
	}
	fmt.Fprintf(c.w, "</classVarDec>\n")
	return nil
}

func (c *compiler) compileType() (string, error) {
	switch {
	case c.gotKeyword(KeywordInt):
		c.advance()
		return "int", nil
	case c.gotKeyword(KeywordChar):
		c.advance()
		return "char", nil
	case c.gotKeyword(KeywordBoolean):
		c.advance()
		return "boolean", nil
	}
	return c.consumeIdentifier()
}

func (c *compiler) compileSubroutine() error {
	fmt.Fprintf(c.w, "<subroutineDec>\n")

	// keyword constructor / function / method
	if err := c.consumeKeyword(KeywordConstructor, KeywordFunction, KeywordMethod); err != nil {
		return err
	}

	// return type or "void"
	if c.gotKeyword(KeywordVoid) {
		c.advance()
	} else {
		if _, err := c.compileType(); err != nil {
			return err
		}
	}

	// name
	subroutineName, err := c.consumeIdentifier()
	if err != nil {
		return err
	}
	_ = subroutineName

	// parameters
	if err := c.compileParameterList(); err != nil {
		return err
	}

	// body
	if err := c.compileSubroutineBody(); err != nil {
		return err
	}

	fmt.Fprintf(c.w, "</subroutineDec>\n")
	c.subroutineTable.Reset()
	return nil
}

func (c *compiler) compileParameterList() error {
	if err := c.consumeSymbol('('); err != nil {
		return err
	}
	fmt.Fprintf(c.w, "<parameterList>\n")
	if c.gotSymbol(')') {
		// empty parameter list
	} else {
		for {
			// TODO define subroutine arguments
			typ, err := c.compileType()
			if err != nil {
				return err
			}
			argumentName, err := c.consumeIdentifier()
			if err != nil {
				return err
			}

			c.subroutineTable.Define(argumentName, typ, SymbolKindArg)
			if c.gotSymbol(')') {
				// reached the end of the parameter list
				break
			} else {
				if err := c.consumeSymbol(','); err != nil {
					return err
				}
			}
		}
	}
	fmt.Fprintf(c.w, "</parameterList>\n")
	c.advance()
	return nil
}

func (c *compiler) compileSubroutineBody() error {
	fmt.Fprintf(c.w, "<subroutineBody>\n")

	// opening {
	if err := c.consumeSymbol('{'); err != nil {
		return err
	}

	// variable declarations
	for c.gotKeyword(KeywordVar) {
		if err := c.compileVarDec(); err != nil {
			return err
		}
	}

	// statements and closing }
	fmt.Fprintf(c.w, "<statements>\n")
	for !c.gotSymbol('}') {
		if err := c.compileStatement(); err != nil {
			return err
		}
	}
	fmt.Fprintf(c.w, "</statements>\n")
	c.advance()

	fmt.Fprintf(c.w, "</subroutineBody>\n")
	return nil
}

func (c *compiler) compileVarDec() error {
	fmt.Fprintf(c.w, "<varDec>\n")
	if err := c.consumeKeyword(KeywordVar); err != nil {
		return err
	}
	typ, err := c.compileType()
	if err != nil {
		return err
	}
	for {
		varName, err := c.consumeIdentifier()
		if err != nil {
			return err
		}
		c.subroutineTable.Define(varName, typ, SymbolKindVar)
		if c.gotSymbol(',') {
			c.advance()
		} else {
			break
		}
	}
	if err := c.consumeSymbol(';'); err != nil {
		return err
	}
	fmt.Fprintf(c.w, "</varDec>\n")
	return nil
}

func (c *compiler) compileStatement() error {
	switch {
	case c.gotKeyword(KeywordLet):
		return c.compileLetStatement()
	case c.gotKeyword(KeywordIf):
		return c.compileIfStatement()
	case c.gotKeyword(KeywordWhile):
		return c.compileWhileStatement()
	case c.gotKeyword(KeywordDo):
		return c.compileDoStatement()
	case c.gotKeyword(KeywordReturn):
		return c.compileReturnStatement()
	}
	return errors.New("expected one of “let”, “if”, “while”, “do”, “return”")
}

func (c *compiler) compileLetStatement() error {
	fmt.Fprintf(c.w, "<letStatement>\n")

	// keyword let
	if err := c.consumeKeyword(KeywordLet); err != nil {
		return err
	}

	// variable name
	varName, err := c.consumeIdentifier()
	if err != nil {
		return err
	}
	_ = varName

	// optional [...]
	if c.gotSymbol('[') {
		c.advance()
		if err := c.compileExpression(); err != nil {
			return err
		}
		if err := c.consumeSymbol(']'); err != nil {
			return err
		}
	}

	// symbol =
	if err := c.consumeSymbol('='); err != nil {
		return err
	}

	// right-hand side
	if err := c.compileExpression(); err != nil {
		return err
	}

	// symbol ;
	if err := c.consumeSymbol(';'); err != nil {
		return err
	}

	fmt.Fprintf(c.w, "</letStatement>\n")
	return nil
}

func (c *compiler) compileIfStatement() error {
	fmt.Fprintf(c.w, "<ifStatement>\n")

	// keyword if
	if err := c.consumeKeyword(KeywordIf); err != nil {
		return err
	}

	// opening (
	if err := c.consumeSymbol('('); err != nil {
		return err
	}

	// expression
	if err := c.compileExpression(); err != nil {
		return err
	}

	// closing )
	if err := c.consumeSymbol(')'); err != nil {
		return err
	}

	// opening {
	if err := c.consumeSymbol('{'); err != nil {
		return err
	}

	// statements and closing }
	fmt.Fprintf(c.w, "<statements>\n")
	for !c.gotSymbol('}') {
		if err := c.compileStatement(); err != nil {
			return err
		}
	}
	fmt.Fprintf(c.w, "</statements>\n")
	c.advance()

	// optional else part
	if c.gotKeyword(KeywordElse) {
		c.advance()

		// opening {
		if err := c.consumeSymbol('{'); err != nil {
			return err
		}

		// statements and closing }
		fmt.Fprintf(c.w, "<statements>\n")
		for !c.gotSymbol('}') {
			if err := c.compileStatement(); err != nil {
				return err
			}
		}
		fmt.Fprintf(c.w, "</statements>\n")
		c.advance()
	}

	fmt.Fprintf(c.w, "</ifStatement>\n")
	return nil
}

func (c *compiler) compileWhileStatement() error {
	fmt.Fprintf(c.w, "<whileStatement>\n")

	// keyword while
	if err := c.consumeKeyword(KeywordWhile); err != nil {
		return err
	}

	// opening (
	if err := c.consumeSymbol('('); err != nil {
		return err
	}

	// expression
	if err := c.compileExpression(); err != nil {
		return err
	}

	// closing )
	if err := c.consumeSymbol(')'); err != nil {
		return err
	}

	// opening {
	if err := c.consumeSymbol('{'); err != nil {
		return err
	}

	// statements and closing }
	fmt.Fprintf(c.w, "<statements>\n")
	for !c.gotSymbol('}') {
		if err := c.compileStatement(); err != nil {
			return err
		}
	}
	fmt.Fprintf(c.w, "</statements>\n")
	c.advance()

	fmt.Fprintf(c.w, "</whileStatement>\n")
	return nil
}

func (c *compiler) compileDoStatement() error {
	fmt.Fprintf(c.w, "<doStatement>\n")
	if err := c.consumeKeyword(KeywordDo); err != nil {
		return err
	}
	identifier, err := c.consumeIdentifier()
	if err != nil {
		return err
	}
	if err := c.compileSubroutineCall(identifier); err != nil {
		return err
	}
	if err := c.consumeSymbol(';'); err != nil {
		return err
	}
	fmt.Fprintf(c.w, "</doStatement>\n")
	return nil
}

func (c *compiler) compileReturnStatement() error {
	fmt.Fprintf(c.w, "<returnStatement>\n")
	if err := c.consumeKeyword(KeywordReturn); err != nil {
		return err
	}
	if c.gotSymbol(';') {
		c.advance()
	} else {
		if err := c.compileExpression(); err != nil {
			return err
		}
		if err := c.consumeSymbol(';'); err != nil {
			return err
		}
	}
	fmt.Fprintf(c.w, "</returnStatement>\n")
	return nil
}

func (c *compiler) compileExpression() error {
	fmt.Fprintf(c.w, "<expression>\n")

	for {
		if err := c.compileTerm(); err != nil {
			return err
		}
		if c.gotSymbol('+', '-', '*', '/', '&', '|', '<', '>', '=') {
			c.advance()
		} else {
			break
		}
	}

	fmt.Fprintf(c.w, "</expression>\n")
	return nil
}

func (c *compiler) compileTerm() error {
	fmt.Fprintf(c.w, "<term>\n")

	if c.got(TokenTypeIntConst, TokenTypeStringConst) {
		// int or string constant
		c.advance()
	} else if c.gotKeyword(KeywordTrue, KeywordFalse, KeywordNull, KeywordThis) {
		// keyword true / false / null / this
		c.advance()
	} else if c.gotSymbol('(') {
		// parenthesized expression
		c.advance()
		if err := c.compileExpression(); err != nil {
			return err
		}
		if err := c.consumeSymbol(')'); err != nil {
			return err
		}
	} else if c.gotSymbol('-', '~') {
		// unary operator
		c.advance()
		if err := c.compileTerm(); err != nil {
			return err
		}
	} else if c.got(TokenTypeIdentifier) {
		// expression starting with an identifier
		identifier, _ := c.gotIdentifier()
		c.advance()
		if c.gotSymbol('[') {
			// array indexing
			c.advance()
			if err := c.compileExpression(); err != nil {
				return err
			}
			if err := c.consumeSymbol(']'); err != nil {
				return err
			}
		} else if c.gotSymbol('(', '.') {
			if err := c.compileSubroutineCall(identifier); err != nil {
				return err
			}
		} else {
			// just a variable
		}
	} else {
		return c.errExpected("term")
	}

	fmt.Fprintf(c.w, "</term>\n")
	return nil
}

func (c *compiler) compileSubroutineCall(identifier string) error {
	_ = identifier
	if c.gotSymbol('.') {
		c.advance()
		_, err := c.consumeIdentifier()
		if err != nil {
			return err
		}
	}
	if err := c.compileExpressionList(); err != nil {
		return err
	}
	return nil
}

func (c *compiler) compileExpressionList() error {
	if err := c.consumeSymbol('('); err != nil {
		return err
	}
	fmt.Fprintf(c.w, "<expressionList>\n")
	if !c.gotSymbol(')') {
		for {
			if err := c.compileExpression(); err != nil {
				return err
			}
			if c.gotSymbol(',') {
				c.advance()
			} else {
				break
			}
		}
	}
	fmt.Fprintf(c.w, "</expressionList>\n")
	if err := c.consumeSymbol(')'); err != nil {
		return err
	}
	return nil
}

func (c *compiler) advance() {
	printToken(c.w, c.t)
	c.atEnd = !c.t.Tokenize()
}

func (c *compiler) consumeKeyword(ks ...Keyword) error {
	err := func() error {
		return c.errExpected(oneOf(ks...))
	}
	if c.atEnd {
		return err()
	}
	if c.t.TokenType() != TokenTypeKeyword {
		return err()
	}
	got := c.t.Keyword()
	for _, k := range ks {
		if got == k {
			c.advance()
			return nil
		}
	}
	return err()
}

func (c *compiler) consumeSymbol(s rune) error {
	if c.atEnd || c.t.TokenType() != TokenTypeSymbol || c.t.Symbol() != s {
		return c.errExpected(fmt.Sprintf("symbol “%c”", s))
	}
	c.advance()
	return nil
}

func (c *compiler) consumeIdentifier() (string, error) {
	if c.t.TokenType() != TokenTypeIdentifier {
		return "", c.errExpected("identifier")
	}
	identifier := c.t.Identifier()
	c.advance()
	return identifier, nil
}

func (c *compiler) got(ts ...TokenType) bool {
	if c.atEnd {
		return false
	}
	got := c.t.TokenType()
	for _, t := range ts {
		if got == t {
			return true
		}
	}
	return false
}

func (c *compiler) gotSymbol(ss ...rune) bool {
	if c.atEnd || c.t.TokenType() != TokenTypeSymbol {
		return false
	}
	got := c.t.Symbol()
	for _, s := range ss {
		if got == s {
			return true
		}
	}
	return false
}

func (c *compiler) gotKeyword(ks ...Keyword) bool {
	if c.atEnd || c.t.TokenType() != TokenTypeKeyword {
		return false
	}
	got := c.t.Keyword()
	for _, k := range ks {
		if got == k {
			return true
		}
	}
	return false
}

func (c *compiler) gotIdentifier() (value string, ok bool) {
	if c.atEnd || c.t.TokenType() != TokenTypeIdentifier {
		return
	}
	return c.t.Identifier(), true
}

func (c *compiler) gotInteger() (value int, ok bool) {
	if c.atEnd || c.t.TokenType() != TokenTypeIntConst {
		return
	}
	return c.t.IntVal(), true
}

func (c *compiler) errExpected(expected string) error {
	var got string
	if c.atEnd {
		got = "end of input"
	} else {
		switch c.t.TokenType() {
		case TokenTypeKeyword:
			got = fmt.Sprintf("keyword “%s”", c.t.Keyword())
		case TokenTypeSymbol:
			got = fmt.Sprintf("“%c”", c.t.Symbol())
		case TokenTypeIdentifier:
			got = fmt.Sprintf("identifier “%s”", c.t.Identifier())
		case TokenTypeIntConst:
			got = fmt.Sprintf("integer constant %d", c.t.IntVal())
		case TokenTypeStringConst:
			got = fmt.Sprintf("string constant “%s”", c.t.StringVal())
		}
	}
	return fmt.Errorf("expected %s; got %s", expected, got)
}

func oneOf[T fmt.Stringer](options ...T) string {
	list := make([]string, len(options))
	for i, o := range options {
		list[i] = fmt.Sprintf("“%s”", o)
	}
	if len(list) == 1 {
		return list[0]
	}
	return fmt.Sprintf("one of %s", strings.Join(list, ", "))
}
