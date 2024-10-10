package ast

import (
	"fmt"
	"strconv"
)

type Parser struct {
	lexer *Lexer
	curr  Token
	next  Token
}

func NewParser(lexer *Lexer) Parser {
	return Parser{
		lexer: lexer,
		curr:  lexer.NextToken(),
		next:  lexer.NextToken(),
	}
}

func (parser *Parser) advance() Token {
	temp := parser.curr
	parser.curr = parser.next
	parser.next = parser.lexer.NextToken()
	return temp
}

func (parser *Parser) peek() TokenType {
	return parser.curr.Type
}

func (parser *Parser) expect(expected TokenType) (Token, error) {
	if parser.curr.Type == expected {
		return parser.advance(), nil
	} else {
		return parser.curr, fmt.Errorf("expected %s", expected)
	}
}

func (parser *Parser) Expr() (Expr, error) {
	switch parser.peek() {
	case TokenEOF:
		return NewExprError("Reached EOF"), fmt.Errorf("reached EOF")
	case TokenError:
		return NewExprError(parser.curr.Lexeme), fmt.Errorf("error")
	case TokenRParens:
		return NewExprError("Unexpected ')'"), fmt.Errorf("unexpected")
	case TokenNumber:
		return parser.number()
	case TokenIdentifier:
		return parser.symbol()
	case TokenString:
		return parser.string()
	case TokenLParens:
		return parser.list()
	default:
		panic("unreachable")
	}
}

func (parser *Parser) number() (Expr, error) {
	token, err := parser.expect(TokenNumber)
	if err != nil {
		return NewExprError("Error"), err
	}

	number, err := strconv.Atoi(token.Lexeme)
	if err != nil {
		return NewExprError("Error"), err
	}

	expr := &Number{Number: number}
	return expr, nil
}

func (parser *Parser) symbol() (Expr, error) {
	token, err := parser.expect(TokenIdentifier)
	if err != nil {
		return NewExprError("Error"), err
	}

	expr := &Symbol{Name: token.Lexeme}
	return expr, nil
}

func (parser *Parser) string() (Expr, error) {
	token, err := parser.expect(TokenString)
	if err != nil {
		return NewExprError("Error"), err
	}

	expr := &String{Contents: token.Lexeme}
	return expr, nil
}

func (parser *Parser) list() (Expr, error) {
	_, err := parser.expect(TokenLParens)
	if err != nil {
		return NewExprError("Error"), err
	}

	elements := make([]Expr, 0)
	for parser.curr.Type != TokenRParens {
		expr, err := parser.Expr()
		if err != nil {
			return expr, err
		}
		elements = append(elements, expr)
	}

	_, err = parser.expect(TokenRParens)
	if err != nil {
		return NewExprError("Error"), err
	}

	return &List{Elements: elements}, nil
}

func (parser *Parser) Program() ([]Expr, error) {
	definitions := []Expr{}

	for {
		t := parser.peek()
		if t == TokenEOF {
			break
		}

		definition, err := parser.list()
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, definition)
	}

	return definitions, nil
}
