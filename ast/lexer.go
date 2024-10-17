package ast

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type TokenType uint8

func (t TokenType) String() string {
	switch t {
	case TokenIdentifier:
		return "TokenIdentifier"
	case TokenNumber:
		return "TokenNumber"
	case TokenString:
		return "TokenString"
	case TokenLParens:
		return "TokenLParens"
	case TokenRParens:
		return "TokenRParens"
	case TokenError:
		return "TokenError"
	case TokenEOF:
		return "TokenEOF"
	}
	panic("Should not happen")
}

const (
	TokenIdentifier TokenType = iota
	TokenNumber
	TokenString
	TokenLParens
	TokenRParens
	TokenLBrace
	TokenRBrace
	TokenError
	TokenEOF
)

type ByteSpan struct {
	Start int
	End   int
}

func (span ByteSpan) String() string {
	return fmt.Sprintf("{%d..%d}", span.Start, span.End)
}

type Token struct {
	Type     TokenType
	Lexeme   string
	ByteSpan ByteSpan
}

type Lexer struct {
	src   string
	start int
	index int

	r bufio.Reader
}

func (lexer *Lexer) save() {
	lexer.start = lexer.index
}

func NewLexer(src string) Lexer {
	r := *bufio.NewReader(strings.NewReader(src))
	return Lexer{
		src:   src,
		start: 0,
		index: 0,
		r:     r,
	}
}

func (lexer *Lexer) peek() (rune, error) {
	r, _, err := lexer.r.ReadRune()
	lexer.r.UnreadRune()
	return r, err
}

func (lexer *Lexer) advance() (rune, error) {
	r, size, err := lexer.r.ReadRune()
	if err != nil {
		return r, err
	}
	lexer.index += size
	return r, err
}

func (lexer *Lexer) advanceWhile(cond func(rune) bool) {
	for {
		r, err := lexer.peek()
		if err != nil {
			return
		}
		if cond(r) {
			lexer.advance()
		} else {
			break
		}
	}
}

func (lexer *Lexer) whitespaces() {
	lexer.advanceWhile(unicode.IsSpace)
}

func (lexer *Lexer) getByteSpan() ByteSpan {
	return ByteSpan{
		Start: lexer.start,
		End:   lexer.index,
	}
}

func (lexer *Lexer) lexeme() string {
	return lexer.src[lexer.start:lexer.index]
}

func (lexer *Lexer) stringToken() Token {
	s := make([]rune, 0)
	for {
		r, err := lexer.peek()
		if err != nil {
			break
		}
		if r != '"' {
			r, _ := lexer.advance()
			s = append(s, r)
		} else {
			break
		}
	}

	r, err := lexer.advance()
	if err != nil || r != '"' {
		return Token{
			Type:     TokenError,
			Lexeme:   string(s),
			ByteSpan: lexer.getByteSpan(),
		}
	}

	return Token{
		Type:     TokenString,
		Lexeme:   string(s),
		ByteSpan: lexer.getByteSpan(),
	}
}

func (lexer *Lexer) NextToken() Token {
	lexer.whitespaces()
	lexer.save()

	r, err := lexer.advance()
	if errors.Is(err, io.EOF) {
		byteSpan := lexer.getByteSpan()
		return Token{
			Type:     TokenEOF,
			Lexeme:   "",
			ByteSpan: byteSpan,
		}
	} else {
		var tokenType TokenType
		switch {
		case r == '(':
			tokenType = TokenLParens
		case r == ')':
			tokenType = TokenRParens
		case r == '{':
			tokenType = TokenLBrace
		case r == '}':
			tokenType = TokenRBrace
		case r == '"':
			return lexer.stringToken()
		case r == '-':
			p, err := lexer.peek()
			if err != nil {
				lexer.advanceWhile(isSymbol)
				tokenType = TokenIdentifier
			} else {
				if unicode.IsDigit(p) {
					lexer.advanceWhile(unicode.IsDigit)
					tokenType = TokenNumber
				} else {
					lexer.advanceWhile(isSymbol)
					tokenType = TokenIdentifier
				}
			}
		case unicode.IsDigit(r):
			lexer.advanceWhile(unicode.IsDigit)
			tokenType = TokenNumber
		case isSymbol(r):
			lexer.advanceWhile(isSymbol)
			tokenType = TokenIdentifier
		default:
			tokenType = TokenError
		}
		lexeme := lexer.lexeme()
		byteSpan := lexer.getByteSpan()
		return Token{
			Type:     tokenType,
			Lexeme:   lexeme,
			ByteSpan: byteSpan,
		}
	}
}

var restricted = map[rune]bool{
	'(':  true,
	')':  true,
	'[':  true,
	']':  true,
	'{':  true,
	'}':  true,
	'`':  true,
	',':  true,
	'\'': true,
}

func isSymbol(r rune) bool {
	isSpace := unicode.IsSpace(r)
	_, found := restricted[r]
	return !found && !isSpace
}
