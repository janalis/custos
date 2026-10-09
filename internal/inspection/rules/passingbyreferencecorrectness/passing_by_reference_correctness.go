package passingbyreferencecorrectness

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// passingByReferenceCorrectness reports non-variable arguments (calls not
// returning by reference, `new`) passed to by-reference parameters.
type passingByReferenceCorrectness struct{}

// Semantic marks the rule as needing the project symbol index.
func (passingByReferenceCorrectness) Semantic()  {}
func (passingByReferenceCorrectness) ID() string { return "PassingByReferenceCorrectness" }
func (passingByReferenceCorrectness) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall, syntax.KStaticCall}
}

func (passingByReferenceCorrectness) Check(ctx *analysis.Context, n syntax.Node) {
	// D1
	switch c := n.(type) {
	case *syntax.FuncCall:
		if _, ok := c.Name.(*syntax.Name); !ok {
			return
		}
		switch ctx.FunctionName(c) { // D2
		case "current", "key":
			if f := ctx.Types().ResolveFunction(c); f == nil || !strings.Contains(strings.TrimPrefix(f.FQN, `\`), `\`) {
				return
			}
		case "array_multisort", "extract":
			// Prefer-ref parameters: temporaries are accepted without a
			// notice (custos).
			if ctx.GlobalFunctionName(c) != "" {
				return
			}
		}
	case *syntax.MethodCall:
		if id, ok := c.Name.(*syntax.Identifier); !ok || id.Value == "" {
			return
		}
	case *syntax.StaticCall:
		if id, ok := c.Name.(*syntax.Identifier); !ok || id.Value == "" {
			return
		}
	}
	list := callArgs(n)
	if list == nil || len(list.Args) == 0 { // D3
		return
	}
	allPlain := true
	for _, a := range list.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok {
			return // first-class callable syntax
		}
		switch arg.Value.(type) {
		case *syntax.Variable:
			continue
		case *syntax.New:
			if ctx.PHP < phpversion.PHP70 {
				continue
			}
		}
		allPlain = false
	}
	if allPlain {
		return
	}
	cal, ok := semanticquery.ResolveCallee(ctx, n) // D4
	if !ok {
		return
	}
	for i, a := range list.Args { // D5
		arg := a.(*syntax.Arg)
		pi := i
		if arg.Name != nil {
			pi = -1
			for j, p := range cal.Params {
				if p.Name == arg.Name.Value {
					pi = j
				}
			}
		}
		if pi < 0 || pi >= len(cal.Params) {
			continue
		}
		if !cal.Params[pi].ByRef || arg.Value == nil {
			continue
		}
		switch v := arg.Value.(type) {
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
			if arg.ByRef {
				continue
			}
			inner, ok := semanticquery.ResolveCallee(ctx, v)
			if ok && !inner.ByRef {
				ctx.ReportNode(v, "Pass a variable here: this parameter is taken by reference.")
			}
		case *syntax.New:
			ctx.ReportNode(v, "Pass a variable here: this parameter is taken by reference.")
		}
	}
}
