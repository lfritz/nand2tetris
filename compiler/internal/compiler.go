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
	syntaxWriter    io.Writer
	vmWriter        *VMWriter
	printMode       bool
	atEnd           bool
	classTable      *SymbolTable
	subroutineTable *SymbolTable
}

func newCompiler(r io.Reader, w io.Writer, printMode bool) *compiler {
	return &compiler{
		t:               NewTokenizer(r),
		syntaxWriter:    discardWriter(w, !printMode),
		vmWriter:        NewVMWriter(discardWriter(w, printMode)),
		printMode:       printMode,
		classTable:      NewSymbolTable(),
		subroutineTable: NewSymbolTable(),
	}
}

func discardWriter(w io.Writer, discard bool) io.Writer {
	if discard {
		return io.Discard
	}
	return w
}

func (c *compiler) lookup(name string) *SymbolTable {
	if _, ok := c.subroutineTable.symbols[name]; ok {
		return c.subroutineTable
	}
	if _, ok := c.classTable.symbols[name]; ok {
		return c.classTable
	}
	return nil
}

func (c *compiler) compileFile() error {
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
	fmt.Fprintf(c.syntaxWriter, "<class>\n")
	if err := c.consumeKeyword(KeywordClass); err != nil {
		return err
	}
	className, err := c.consumeIdentifier()
	if err != nil {
		return err
	}
	c.printIdentifier(className, "class", -1, "declaration")
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
	fmt.Fprintf(c.syntaxWriter, "</class>\n")
	c.classTable.Reset()
	return nil
}

func (c *compiler) compileClassVarDec() error {
	fmt.Fprintf(c.syntaxWriter, "<classVarDec>\n")
	kind := SymbolKindField
	category := "field"
	if c.gotKeyword(KeywordStatic) {
		kind = SymbolKindStatic
		category = "static"
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
		index := c.classTable.Define(name, typ, kind)
		c.printIdentifier(name, category, index, "declaration")
		if c.gotSymbol(',') {
			c.advance()
		} else {
			break
		}
	}
	if err := c.consumeSymbol(';'); err != nil {
		return err
	}
	fmt.Fprintf(c.syntaxWriter, "</classVarDec>\n")
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
	fmt.Fprintf(c.syntaxWriter, "<subroutineDec>\n")

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

	fmt.Fprintf(c.syntaxWriter, "</subroutineDec>\n")
	c.subroutineTable.Reset()
	return nil
}

func (c *compiler) compileParameterList() error {
	if err := c.consumeSymbol('('); err != nil {
		return err
	}
	fmt.Fprintf(c.syntaxWriter, "<parameterList>\n")
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

			index := c.subroutineTable.Define(argumentName, typ, SymbolKindArg)
			c.printIdentifier(argumentName, "arg", index, "declaration")
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
	fmt.Fprintf(c.syntaxWriter, "</parameterList>\n")
	c.advance()
	return nil
}

func (c *compiler) compileSubroutineBody() error {
	fmt.Fprintf(c.syntaxWriter, "<subroutineBody>\n")

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
	fmt.Fprintf(c.syntaxWriter, "<statements>\n")
	for !c.gotSymbol('}') {
		if err := c.compileStatement(); err != nil {
			return err
		}
	}
	fmt.Fprintf(c.syntaxWriter, "</statements>\n")
	c.advance()

	fmt.Fprintf(c.syntaxWriter, "</subroutineBody>\n")
	return nil
}

func (c *compiler) compileVarDec() error {
	fmt.Fprintf(c.syntaxWriter, "<varDec>\n")
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
		index := c.subroutineTable.Define(varName, typ, SymbolKindVar)
		c.printIdentifier(varName, "var", index, "declaration")
		if c.gotSymbol(',') {
			c.advance()
		} else {
			break
		}
	}
	if err := c.consumeSymbol(';'); err != nil {
		return err
	}
	fmt.Fprintf(c.syntaxWriter, "</varDec>\n")
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
	fmt.Fprintf(c.syntaxWriter, "<letStatement>\n")

	// keyword let
	if err := c.consumeKeyword(KeywordLet); err != nil {
		return err
	}

	// variable name
	varName, err := c.consumeIdentifier()
	if err != nil {
		return err
	}
	c.printIdentifierUse(varName)

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

	fmt.Fprintf(c.syntaxWriter, "</letStatement>\n")
	return nil
}

func (c *compiler) compileIfStatement() error {
	fmt.Fprintf(c.syntaxWriter, "<ifStatement>\n")

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
	fmt.Fprintf(c.syntaxWriter, "<statements>\n")
	for !c.gotSymbol('}') {
		if err := c.compileStatement(); err != nil {
			return err
		}
	}
	fmt.Fprintf(c.syntaxWriter, "</statements>\n")
	c.advance()

	// optional else part
	if c.gotKeyword(KeywordElse) {
		c.advance()

		// opening {
		if err := c.consumeSymbol('{'); err != nil {
			return err
		}

		// statements and closing }
		fmt.Fprintf(c.syntaxWriter, "<statements>\n")
		for !c.gotSymbol('}') {
			if err := c.compileStatement(); err != nil {
				return err
			}
		}
		fmt.Fprintf(c.syntaxWriter, "</statements>\n")
		c.advance()
	}

	fmt.Fprintf(c.syntaxWriter, "</ifStatement>\n")
	return nil
}

func (c *compiler) compileWhileStatement() error {
	fmt.Fprintf(c.syntaxWriter, "<whileStatement>\n")

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
	fmt.Fprintf(c.syntaxWriter, "<statements>\n")
	for !c.gotSymbol('}') {
		if err := c.compileStatement(); err != nil {
			return err
		}
	}
	fmt.Fprintf(c.syntaxWriter, "</statements>\n")
	c.advance()

	fmt.Fprintf(c.syntaxWriter, "</whileStatement>\n")
	return nil
}

func (c *compiler) compileDoStatement() error {
	fmt.Fprintf(c.syntaxWriter, "<doStatement>\n")
	if err := c.consumeKeyword(KeywordDo); err != nil {
		return err
	}
	if err := c.compileExpression(); err != nil {
		return err
	}
	if err := c.consumeSymbol(';'); err != nil {
		return err
	}
	c.vmWriter.WritePop(SegmentTemp, 0)
	fmt.Fprintf(c.syntaxWriter, "</doStatement>\n")
	return nil
}

func (c *compiler) compileReturnStatement() error {
	fmt.Fprintf(c.syntaxWriter, "<returnStatement>\n")
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
	fmt.Fprintf(c.syntaxWriter, "</returnStatement>\n")
	c.vmWriter.WriteReturn()
	return nil
}

func (c *compiler) compileExpression() error {
	fmt.Fprintf(c.syntaxWriter, "<expression>\n")

	var operator rune
	for {
		if err := c.compileTerm(); err != nil {
			return err
		}
		switch operator {
		case '+':
			c.vmWriter.WriteArithmetic(CommandAdd)
		case '-':
			c.vmWriter.WriteArithmetic(CommandSub)
		case '*':
			c.vmWriter.WriteCall("Math.multiply", 2)
		case '/':
			c.vmWriter.WriteCall("Math.divide", 2)
		case '&':
			c.vmWriter.WriteArithmetic(CommandAnd)
		case '|':
			c.vmWriter.WriteArithmetic(CommandOr)
		case '<':
			c.vmWriter.WriteArithmetic(CommandLt)
		case '>':
			c.vmWriter.WriteArithmetic(CommandGt)
		case '=':
			c.vmWriter.WriteArithmetic(CommandEq)
		}
		if c.gotSymbol('+', '-', '*', '/', '&', '|', '<', '>', '=') {
			operator = c.t.Symbol()
			c.advance()
		} else {
			break
		}
	}

	fmt.Fprintf(c.syntaxWriter, "</expression>\n")
	return nil
}

func (c *compiler) compileTerm() error {
	fmt.Fprintf(c.syntaxWriter, "<term>\n")

	if c.got(TokenTypeIntConst) {
		// int constant
		c.vmWriter.WritePush(SegmentConstant, c.t.intVal)
		c.advance()
	} else if c.got(TokenTypeStringConst) {
		// string constant
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
			c.printIdentifierUse(identifier)
			c.advance()
			if err := c.compileExpression(); err != nil {
				return err
			}
			if err := c.consumeSymbol(']'); err != nil {
				return err
			}
		} else if c.gotSymbol('(', '.') {
			// subroutine call
			if err := c.compileSubroutineCall(identifier); err != nil {
				return err
			}
		} else {
			// just a variable
			segment, index := c.translateVariable(identifier)
			c.vmWriter.WritePush(segment, index)
			c.printIdentifierUse(identifier)
		}
	} else {
		return c.errExpected("term")
	}

	fmt.Fprintf(c.syntaxWriter, "</term>\n")
	return nil
}

func (c *compiler) translateVariable(name string) (Segment, int) {
	table := c.lookup(name)
	kind := table.KindOf(name)
	index := table.IndexOf(name)
	switch kind {
	case SymbolKindStatic:
		// TODO
	case SymbolKindField:
		// TODO
	case SymbolKindArg:
		// TODO
	case SymbolKindVar:
		return SegmentLocal, index
	}
	return 0, 0
}

func (c *compiler) compileSubroutineCall(identifier string) error {
	name := identifier
	if c.gotSymbol('.') {
		c.printIdentifier(identifier, "class", -1, "use")
		c.advance()
		subroutineName, err := c.consumeIdentifier()
		if err != nil {
			return err
		}
		c.printIdentifier(subroutineName, "subroutine", -1, "use")
		name = name + "." + subroutineName
	} else {
		c.printIdentifier(identifier, "subroutine", -1, "use")
	}
	nArgs, err := c.compileExpressionList()
	if err != nil {
		return err
	}
	c.vmWriter.WriteCall(name, nArgs)
	return nil
}

func (c *compiler) compileExpressionList() (int, error) {
	count := 0
	if err := c.consumeSymbol('('); err != nil {
		return 0, err
	}
	fmt.Fprintf(c.syntaxWriter, "<expressionList>\n")
	if !c.gotSymbol(')') {
		for {
			if err := c.compileExpression(); err != nil {
				return 0, err
			}
			count++
			if c.gotSymbol(',') {
				c.advance()
			} else {
				break
			}
		}
	}
	fmt.Fprintf(c.syntaxWriter, "</expressionList>\n")
	if err := c.consumeSymbol(')'); err != nil {
		return 0, err
	}
	return count, nil
}

func (c *compiler) printIdentifierUse(name string) {
	table := c.lookup(name)
	kind := table.KindOf(name)
	index := table.IndexOf(name)

	var category string
	switch kind {
	case SymbolKindStatic:
		category = "static"
	case SymbolKindField:
		category = "field"
	case SymbolKindArg:
		category = "arg"
	case SymbolKindVar:
		category = "var"
	}

	fmt.Fprintf(c.syntaxWriter, "<identifier>\n")
	fmt.Fprintf(c.syntaxWriter, "<name> %s </name>\n", name)
	fmt.Fprintf(c.syntaxWriter, "<category> %s </category>\n", category)
	fmt.Fprintf(c.syntaxWriter, "<index> %d </index>\n", index)
	fmt.Fprintf(c.syntaxWriter, "<usage> use </usage>\n")
	fmt.Fprintf(c.syntaxWriter, "</identifier>\n")
}

func (c *compiler) printIdentifier(name, category string, index int, usage string) {
	fmt.Fprintf(c.syntaxWriter, "<identifier>\n")
	fmt.Fprintf(c.syntaxWriter, "<name> %s </name>\n", name)
	fmt.Fprintf(c.syntaxWriter, "<category> %s </category>\n", category)
	if index >= 0 {
		fmt.Fprintf(c.syntaxWriter, "<index> %d </index>\n", index)
	}
	fmt.Fprintf(c.syntaxWriter, "<usage> %s </usage>\n", usage)
	fmt.Fprintf(c.syntaxWriter, "</identifier>\n")
}

func (c *compiler) advance() {
	switch c.t.TokenType() {
	case TokenTypeKeyword:
		fmt.Fprintf(c.syntaxWriter, "<keyword> %s </keyword>\n", c.t.Keyword().String())
	case TokenTypeSymbol:
		fmt.Fprintf(c.syntaxWriter, "<symbol> %s </symbol>\n", symbolToXML(c.t.Symbol()))
	case TokenTypeIdentifier:
		// identifiers are printed with printIdentifier and printIdentifierUse
	case TokenTypeIntConst:
		fmt.Fprintf(c.syntaxWriter, "<integerConstant> %d </integerConstant>\n", c.t.IntVal())
	case TokenTypeStringConst:
		fmt.Fprintf(c.syntaxWriter, "<stringConstant> %s </stringConstant>\n", c.t.StringVal())
	}

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
