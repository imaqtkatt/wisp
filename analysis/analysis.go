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

func AnalyzeProgram(exprs []ast.Expr) []Form {
	analyzed := []Form{}
	for _, expr := range exprs {
		analyzed = append(analyzed, Analyze(expr))
	}
	return analyzed
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
		if dispatch, found := formsTable[hd.Name]; found {
			anal.output = dispatch(rest)
		} else {
			anal.output = callForm(head, rest)
		}
	default:
		anal.output = callForm(head, rest)
	}
}

func (anal *Analyzer) VisitObject(object *ast.Object) {
	entries := map[Form]Form{}

	for key, value := range object.Entries {
		entries[Analyze(key)] = Analyze(value)
	}

	anal.output = Object{Entries: entries}
}

type FuncAnalyzer func([]ast.Expr) Form

var formsTable map[string]FuncAnalyzer = map[string]FuncAnalyzer{
	"do":    doForm,
	"def":   defForm,
	"defun": defunForm,
	"echo":  echoForm,
	"fun":   funForm,
	"if":    ifForm,
	"let":   letForm,
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

func defunForm(rest []ast.Expr) Form {
	restLen := len(rest)
	if restLen != 3 {
		return FormError{Message: "Expected to have 2 more expressions"}
	}

	sym := rest[0]

	name, err := assertSymbol(sym)
	if err != nil {
		return FormError{Message: "Expected symbol name"}
	}

	list, err := assertList(rest[1])
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

	body := rest[2]
	analyzedBody := Analyze(body)

	return Defun{
		Name:       name,
		Parameters: parametersList,
		Body:       analyzedBody,
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

func ifForm(exprs []ast.Expr) Form {
	exprsLen := len(exprs)
	if exprsLen != 3 {
		return FormError{Message: "Expected three forms"}
	}

	condition := Analyze(exprs[0])
	then := Analyze(exprs[1])
	else_ := Analyze(exprs[2])

	return If{
		Condition: condition,
		Then:      then,
		Else:      else_,
	}
}
