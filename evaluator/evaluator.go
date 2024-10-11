package evaluator

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"wisp/analysis"
)

type EvaluatorContext struct {
	variables map[string]Value
	closing   *EvaluatorContext
	w         io.Writer
}

func newEvaluatorContext(parent *EvaluatorContext) *EvaluatorContext {
	return &EvaluatorContext{
		variables: map[string]Value{},
		closing:   parent,
		w:         parent.w,
	}
}

func (ctx *EvaluatorContext) fetch(name string) (Value, error) {
	curr := ctx
	for {
		if curr == nil {
			return nil, fmt.Errorf("unbound variable '%s'", name)
		}

		value, found := curr.variables[name]
		if !found {
			curr = curr.closing
			continue
		}

		return value, nil
	}
}

func (ctx *EvaluatorContext) def(name string, value Value) {
	ctx.variables[name] = value
}

func (ctx *EvaluatorContext) EvalProgram(anal []analysis.Form) (Value, error) {
	var returnValue Value = NIL
	for _, form := range anal {
		if value, err := ctx.Eval(form); err != nil {
			return nil, err
		} else {
			returnValue = value
		}
	}
	return returnValue, nil
}

func (ctx *EvaluatorContext) Eval(anal analysis.Form) (Value, error) {
	switch form := anal.(type) {
	case analysis.FormError:
		return nil, fmt.Errorf(form.Message)

	case analysis.Number:
		return ValueNumber{form.Value}, nil

	case analysis.String:
		return ValueString{form.Contents}, nil

	case analysis.Symbol:
		return ctx.fetch(form.Name)

	case analysis.Def:
		body, err := ctx.Eval(form.Body)
		if err != nil {
			return nil, err
		}
		ctx.def(form.Name, body)
		return NIL, nil

	case analysis.Defun:
		fun := ValueClosure{
			ctx:        newEvaluatorContext(ctx),
			parameters: form.Parameters,
			body:       form.Body,
		}
		ctx.def(form.Name, fun)
		return fun, nil

	case analysis.Let:
		letCtx := newEvaluatorContext(ctx)
		for _, bind := range form.Binds {
			value, err := letCtx.Eval(bind.Value)
			if err != nil {
				return nil, err
			}

			letCtx.def(bind.Symbol, value)
		}
		return letCtx.Eval(form.Body)

	case analysis.Do:
		var returnValue Value = ValueNil{}
		for _, form := range form.Forms {
			value, err := ctx.Eval(form)
			if err != nil {
				return nil, err
			}
			returnValue = value
		}
		return returnValue, nil

	case analysis.Fun:
		return ValueClosure{
			ctx:        newEvaluatorContext(ctx),
			parameters: form.Parameters,
			body:       form.Body,
		}, nil

	case analysis.Call:
		fun, err := ctx.Eval(form.Callee)
		if err != nil {
			return nil, err
		}
		if fun.IsCallable() {
			arguments := []Value{}

			for _, expr := range form.Arguments {
				value, err := ctx.Eval(expr)
				if err != nil {
					return nil, err
				}
				arguments = append(arguments, value)
			}

			return fun.Call(arguments)
		} else {
			return nil, fmt.Errorf("head is not a callable value")
		}

	case analysis.Echo:
		for _, form := range form.Forms {
			value, err := ctx.Eval(form)
			if err != nil {
				return nil, err
			}
			_, err = fmt.Fprint(ctx.w, value.String())
			if err != nil {
				return nil, err
			}
		}
		return NIL, nil
	}
	panic("unreachable")
}

func Eval(anal analysis.Form) (Value, error) {
	return defaultCtx.Eval(anal)
}

func (ctx *EvaluatorContext) defun(name string, fun func([]Value) (Value, error)) {
	ctx.variables[name] = ValueFun{Fun: fun}
}

var defaultCtx *EvaluatorContext = NewContextWithWriter(os.Stdout)

func NewContextWithWriter(w io.Writer) *EvaluatorContext {
	ctx := &EvaluatorContext{
		variables: map[string]Value{},
		closing:   nil,
		w:         w,
	}

	ctx.def("nil", NIL)
	ctx.def("true", TRUE)
	ctx.def("false", FALSE)

	ctx.defun("inc", inc)
	ctx.defun("nil?", isNil)
	ctx.defun("atoi", atoi)
	ctx.defun("+", add)

	return ctx
}

type ArityError struct {
	Arity int
}

func (e *ArityError) Error() string {
	return fmt.Sprintf("arity error, expected %d argument(s)", e.Arity)
}

type TypeError struct{}

func (e *TypeError) Error() string {
	return "type error"
}

func inc(arguments []Value) (Value, error) {
	if len(arguments) != 1 {
		return nil, &ArityError{Arity: 1}
	}

	switch v := arguments[0].(type) {
	case ValueNumber:
		result := v.Number + 1
		return ValueNumber{Number: result}, nil
	default:
		return nil, &TypeError{}
	}
}

func isNil(arguments []Value) (Value, error) {
	if len(arguments) != 1 {
		return nil, &ArityError{Arity: 1}
	}

	switch arguments[0].(type) {
	case ValueNil:
		return ValueNumber{Number: 1}, nil
	default:
		return ValueNumber{Number: 0}, nil
	}
}

func atoi(arguments []Value) (Value, error) {
	if len(arguments) != 1 {
		return nil, &ArityError{Arity: 1}
	}

	switch v := arguments[0].(type) {
	case ValueString:
		number, err := strconv.Atoi(v.Contents)
		if err != nil {
			return nil, err
		}
		return ValueNumber{Number: number}, nil
	default:
		return nil, &TypeError{}
	}
}

func add(arguments []Value) (Value, error) {
	acc := 0

	if len(arguments) < 1 {
		return nil, &ArityError{Arity: 1}
	}

	for _, value := range arguments {
		switch v := value.(type) {
		case ValueNumber:
			acc += v.Number
		default:
			return nil, &TypeError{}
		}
	}

	return ValueNumber{Number: acc}, nil
}
