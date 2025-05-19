package internal

import (
	"fmt"
)

type TokenType int

const (
	TokenTypeKeyword TokenType = iota
	TokenTypeSymbol
	TokenTypeIdentifier
	TokenTypeIntConst
	TokenTypeStringConst
)

func (t TokenType) String() string {
	switch t {
	case TokenTypeKeyword:
		return "Keyword"
	case TokenTypeSymbol:
		return "Symbol"
	case TokenTypeIdentifier:
		return "Identifier"
	case TokenTypeIntConst:
		return "IntConst"
	case TokenTypeStringConst:
		return "StringConst"
	}
	return fmt.Sprintf("<token type %d>", t)
}

type Keyword int

const (
	KeywordClass Keyword = iota
	KeywordMethod
	KeywordFunction
	KeywordConstructor
	KeywordInt
	KeywordBoolean
	KeywordChar
	KeywordVoid
	KeywordVar
	KeywordStatic
	KeywordField
	KeywordLet
	KeywordDo
	KeywordIf
	KeywordElse
	KeywordWhile
	KeywordReturn
	KeywordTrue
	KeywordFalse
	KeywordNull
	KeywordThis
)

func (k Keyword) String() string {
	switch k {
	case KeywordClass:
		return "class"
	case KeywordMethod:
		return "method"
	case KeywordFunction:
		return "function"
	case KeywordConstructor:
		return "constructor"
	case KeywordInt:
		return "int"
	case KeywordBoolean:
		return "boolean"
	case KeywordChar:
		return "char"
	case KeywordVoid:
		return "void"
	case KeywordVar:
		return "var"
	case KeywordStatic:
		return "static"
	case KeywordField:
		return "field"
	case KeywordLet:
		return "let"
	case KeywordDo:
		return "do"
	case KeywordIf:
		return "if"
	case KeywordElse:
		return "else"
	case KeywordWhile:
		return "while"
	case KeywordReturn:
		return "return"
	case KeywordTrue:
		return "true"
	case KeywordFalse:
		return "false"
	case KeywordNull:
		return "null"
	case KeywordThis:
		return "this"
	}
	return fmt.Sprintf("<keyword %d>", k)
}

var keywords = map[string]Keyword{
	"class":       KeywordClass,
	"method":      KeywordMethod,
	"function":    KeywordFunction,
	"constructor": KeywordConstructor,
	"int":         KeywordInt,
	"boolean":     KeywordBoolean,
	"char":        KeywordChar,
	"void":        KeywordVoid,
	"var":         KeywordVar,
	"static":      KeywordStatic,
	"field":       KeywordField,
	"let":         KeywordLet,
	"do":          KeywordDo,
	"if":          KeywordIf,
	"else":        KeywordElse,
	"while":       KeywordWhile,
	"return":      KeywordReturn,
	"true":        KeywordTrue,
	"false":       KeywordFalse,
	"null":        KeywordNull,
	"this":        KeywordThis,
}
