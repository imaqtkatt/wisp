package evaluator

import (
	"fmt"
	"wisp/analysis"
)

var (
	NIL   = ValueNil{}
	TRUE  = ValueNumber{Number: 1}
	FALSE = ValueNumber{Number: 0}
)

type Value interface {
	IsCallable() bool
	Call(arguments []Value) (Value, error)

	Compare(other Value) bool

	IsTruthy() bool

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

func (ValueNil) IsTruthy() bool {
	return false
}

func (ValueNil) Compare(other Value) bool {
	_, ok := other.(ValueNil)
	return ok
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

func (n ValueNumber) IsTruthy() bool {
	return n.Number != 0
}

func (n ValueNumber) Compare(other Value) bool {
	m, ok := other.(ValueNumber)
	if !ok {
		return false
	}
	return n.Number == m.Number
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

func (ValueString) IsTruthy() bool {
	// consider empty strings as false?
	return true
}

func (s ValueString) Compare(other Value) bool {
	o, ok := other.(ValueString)
	if !ok {
		return false
	}
	return s.Contents == o.Contents
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

func (ValueFun) IsTruthy() bool {
	return true
}

func (ValueFun) Compare(other Value) bool {
	// can't compare functions
	return false
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

func (ValueClosure) IsTruthy() bool {
	return true
}

func (ValueClosure) Compare(other Value) bool {
	// can't compare closures
	return false
}

type ValueObject struct {
	Entries map[Value]Value
}

func (ValueObject) IsTruthy() bool {
	return true
}

func (obj ValueObject) String() string {
	// probably wrong
	return fmt.Sprintf("{%v}", obj.Entries)
}

func (ValueObject) IsCallable() bool {
	return true
}

func (obj ValueObject) Call(arguments []Value) (Value, error) {
	if len(arguments) != 1 {
		panic("arity error")
	}

	toFind := arguments[0]
	value, found := obj.Entries[toFind]
	if !found {
		return NIL, nil
	}
	return value, nil
}

func (obj ValueObject) Compare(other Value) bool {
	o, ok := other.(ValueObject)
	if !ok {
		return false
	}

	for key, val := range obj.Entries {
		gotVal, found := o.Entries[key]
		if !found {
			return false
		}
		if !gotVal.Compare(val) {
			return false
		}
	}

	return true
}
