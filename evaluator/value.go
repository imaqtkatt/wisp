package evaluator

import (
	"fmt"
	"wisp/analysis"
)

var (
	NIL = ValueNil{}
)

type Value interface {
	IsCallable() bool
	Call(arguments []Value) (Value, error)

	String() string
}

type ValueNil struct{}

func (ValueNil) String() string {
	return "nil"
}

func (ValueNil) IsCallable() bool {
	return false
}

func (ValueNil) Call(arguments []Value) (Value, error) {
	panic("Nil is not a callable value")
}

type ValueNumber struct {
	Number int
}

func (n ValueNumber) String() string {
	return fmt.Sprintf("%d", n.Number)
}

func (ValueNumber) IsCallable() bool {
	return false
}

func (ValueNumber) Call(arguments []Value) (Value, error) {
	panic("Number is not a callable value")
}

type ValueString struct {
	Contents string
}

func (s ValueString) String() string {
	return s.Contents
}

func (ValueString) IsCallable() bool {
	return false
}

func (ValueString) Call(arguments []Value) (Value, error) {
	panic("String is not a callable value")
}

type ValueFun struct {
	Fun func([]Value) (Value, error)
}

func (ValueFun) String() string {
	return "<fun>"
}

func (ValueFun) IsCallable() bool {
	return true
}

func (fun ValueFun) Call(arguments []Value) (Value, error) {
	return fun.Fun(arguments)
}

type ValueClosure struct {
	ctx        *EvaluatorContext
	parameters []string
	body       analysis.Form
}

func (ValueClosure) String() string {
	return "<closure>"
}

func (ValueClosure) IsCallable() bool {
	return true
}

func (closure ValueClosure) Call(arguments []Value) (Value, error) {
	arity := len(closure.parameters)
	if len(arguments) != arity {
		panic("arity error")
	}

	for i, param := range closure.parameters {
		closure.ctx.def(param, arguments[i])
	}

	body, err := closure.ctx.Eval(closure.body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
