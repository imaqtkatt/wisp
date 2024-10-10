package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"wisp/analysis"
	"wisp/ast"
	"wisp/evaluator"
)

func main() {
	script := `
(do
  (def id (fun (x) x))
  (id nil)
  (let (x 1
  		y (atoi "3"))
	(echo "result is: " "<p>" (+ x y) "</p>")))
`

	handler, err := scriptHandler(script)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	err = http.ListenAndServe(":80", mux)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Server closed")
	} else {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}

func scriptHandler(script string) (http.HandlerFunc, error) {
	lexer := ast.NewLexer(script)
	parser := ast.NewParser(&lexer)
	expr, err := parser.Expr()
	if err != nil {
		return nil, err
	}

	anal := analysis.Analyze(expr)

	fun := func(w http.ResponseWriter, r *http.Request) {
		buf := bufio.NewWriter(w)
		ctx := evaluator.NewContextWithWriter(buf)
		_, err = ctx.Eval(anal)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte("<html>")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := buf.Flush(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err := w.Write([]byte("</html>")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	return fun, nil
}
