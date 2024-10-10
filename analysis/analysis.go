package analysis

import (
	"fmt"
	"wisp/ast"
)

type Analyzer struct {
	output Form
}

func Analyze(expr ast.Expr) Form {
	analyzer := Analyzer{}
	expr.Accept(&analyzer)
	return analyzer.output
}

func (anal *Analyzer) VisitSymbol(symbol *ast.Symbol) {
	anal.output = Symbol{Name: symbol.Name}
}

func (anal *Analyzer) VisitNumber(number *ast.Number) {
	anal.output = Number{Value: number.Number}
}

func (anal *Analyzer) VisitString(s *ast.String) {
	anal.output = String{Contents: s.Contents}
}

func (anal *Analyzer) VisitList(list *ast.List) {
	listLen := len(list.Elements)

	if listLen < 1 {
		anal.output = FormError{Message: "Expression should have one or more expressions"}
		return
	}

	head := list.Elements[0]
	rest := list.Elements[1:]

	switch hd := head.(type) {
	case *ast.Symbol:
		switch hd.Name {
		case "let":
			anal.output = letForm(rest)
		case "do":
			anal.output = doForm(rest)
		case "def":
			anal.output = defForm(rest)
		case "fun":
			anal.output = funForm(rest)
		case "echo":
			anal.output = echoForm(rest)
		default:
			anal.output = callForm(head, rest)
		}
	default:
		anal.output = callForm(head, rest)
	}
}

func letForm(exprs []ast.Expr) Form {
	exprsLen := len(exprs)
	if exprsLen != 2 {
		return FormError{Message: "Expected 2 expressions"}
	}

	bindsSeq, err := assertList(exprs[0])
	if err != nil {
		return FormError{Message: "Expected list"}
	}

	letBinds := []BindPair{}
	bindsSeqLen := len(bindsSeq)
	for i := 0; i < bindsSeqLen; i += 2 {
		symIdx := i
		valIdx := i + 1

		if valIdx >= bindsSeqLen {
			return FormError{Message: "Expected bind pair value"}
		}

		sym, err := assertSymbol(bindsSeq[symIdx])
		if err != nil {
			return FormError{Message: "Expected bind pair symbol"}
		}

		value := Analyze(bindsSeq[valIdx])

		letBinds = append(letBinds, BindPair{
			Symbol: sym,
			Value:  value,
		})
	}

	body := Analyze(exprs[1])

	return Let{
		Binds: letBinds,
		Body:  body,
	}
}

func funForm(rest []ast.Expr) Form {
	restLen := len(rest)
	if restLen != 2 {
		return FormError{Message: "Expected 2 expressions"}
	}

	list, err := assertList(rest[0])
	if err != nil {
		return FormError{Message: "Expected list"}
	}

	parametersList := []string{}
	for _, expr := range list {
		param, err := assertSymbol(expr)
		if err != nil {
			return FormError{Message: "Expected symbol"}
		}
		parametersList = append(parametersList, param)
	}

	body := Analyze(rest[1])

	return Fun{
		Parameters: parametersList,
		Body:       body,
	}
}

func doForm(exprs []ast.Expr) Form {
	exprsLen := len(exprs)
	if exprsLen < 1 {
		return FormError{Message: "Empty do form"}
	}

	analyzed := []Form{}

	for _, expr := range exprs {
		analyzed = append(analyzed, Analyze(expr))
	}

	return Do{Forms: analyzed}
}

func defForm(rest []ast.Expr) Form {
	restLen := len(rest)
	if restLen != 2 {
		return FormError{Message: "Expected to have 2 more expressions"}
	}

	sym := rest[0]

	name, err := assertSymbol(sym)
	if err != nil {
		return FormError{Message: "Expected symbol name"}
	}

	body := rest[1]
	analyzedBody := Analyze(body)

	return Def{
		Name: name,
		Body: analyzedBody,
	}
}

func assertSymbol(expr ast.Expr) (string, error) {
	switch v := expr.(type) {
	case *ast.Symbol:
		return v.Name, nil
	default:
		return "", fmt.Errorf("not a symbol")
	}
}

func assertList(expr ast.Expr) ([]ast.Expr, error) {
	switch v := expr.(type) {
	case *ast.List:
		return v.Elements, nil
	default:
		return nil, fmt.Errorf("not a list")
	}
}

func callForm(head ast.Expr, tail []ast.Expr) Form {
	analyzedHead := Analyze(head)

	analyzedTail := []Form{}
	for _, expr := range tail {
		analyzedTail = append(analyzedTail, Analyze(expr))
	}

	return Call{
		Callee:    analyzedHead,
		Arguments: analyzedTail,
	}
}

func echoForm(exprs []ast.Expr) Form {
	forms := []Form{}
	for _, expr := range exprs {
		forms = append(forms, Analyze(expr))
	}

	return Echo{Forms: forms}
}
