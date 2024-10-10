package analysis

import "fmt"

type Form interface {
	String() string
}

type FormError struct {
	Message string
}

func (e FormError) String() string {
	return e.Message
}

type Symbol struct {
	Name string
}

func (symbol Symbol) String() string {
	return fmt.Sprintf("Symbol(%+v)", symbol.Name)
}

type Number struct {
	Value int
}

func (number Number) String() string {
	return fmt.Sprintf("Number(%+v)", number.Value)
}

type String struct {
	Contents string
}

func (s String) String() string {
	return fmt.Sprintf("String(%+v)", s.Contents)
}

type Call struct {
	Callee    Form
	Arguments []Form
}

func (call Call) String() string {
	return fmt.Sprintf("Call(%+v, %+v)", call.Callee, call.Arguments)
}

type Do struct {
	Forms []Form
}

func (do Do) String() string {
	return fmt.Sprintf("Do(%+v)", do.Forms)
}

type Def struct {
	Name string
	Body Form
}

func (def Def) String() string {
	return fmt.Sprintf("Def(%+v, %+v)", def.Name, def.Body)
}

type Fun struct {
	Parameters []string
	Body       Form
}

func (fun Fun) String() string {
	return fmt.Sprintf("Fun(%+v, %+v)", fun.Parameters, fun.Body)
}

type Let struct {
	Binds []BindPair
	Body  Form
}

type BindPair struct {
	Symbol string
	Value  Form
}

func (let Let) String() string {
	return fmt.Sprintf("Let(%+v, %+v)", let.Binds, let.Body)
}

type Echo struct {
	Forms []Form
}

func (echo Echo) String() string {
	return fmt.Sprintf("Echo(%+v)", echo.Forms)
}
