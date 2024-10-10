package ast

import "fmt"

type Expr interface {
	String() string
	Accept(visitor ExprVisitor)
}

type ExprError struct {
	Message string
}

func NewExprError(message string) *ExprError {
	return &ExprError{Message: message}
}

func (e *ExprError) Accept(visitor ExprVisitor) {
	panic(fmt.Sprintf("Error: %s", e.Message))
}

func (e ExprError) String() string {
	return e.Message
}

type Symbol struct {
	Name string
}

func (s Symbol) String() string {
	return fmt.Sprintf("Symbol(%v)", s.Name)
}

func (s *Symbol) Accept(visitor ExprVisitor) {
	visitor.VisitSymbol(s)
}

type String struct {
	Contents string
}

func (s String) String() string {
	return fmt.Sprintf("String(%v)", s.Contents)
}

func (s *String) Accept(visitor ExprVisitor) {
	visitor.VisitString(s)
}

type Number struct {
	Number int
}

func (n Number) String() string {
	return fmt.Sprintf("Number(%v)", n.Number)
}

func (n *Number) Accept(visitor ExprVisitor) {
	visitor.VisitNumber(n)
}

type List struct {
	Elements []Expr
}

func (e *List) Accept(visitor ExprVisitor) {
	visitor.VisitList(e)
}

func (e List) String() string {
	return fmt.Sprintf("List(%v)", e.Elements)
}

type ExprVisitor interface {
	VisitSymbol(symbol *Symbol)
	VisitNumber(number *Number)
	VisitString(string *String)
	VisitList(list *List)
}
