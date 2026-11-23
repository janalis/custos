package flow

import (
	"sync"

	"github.com/janalis/custos/internal/php/syntax"
)

type expressionSizesKey struct{}

// expressionSizes estimates the two map capacities separately. Lexical names,
// declaration headers and plain assignment targets are not evaluated values.
// Control wrappers lower to their tests and bodies rather than transfer nodes.
// These are allocation hints only: both maps grow for any additional entries.
func (e *Env) expressionSizes(scope syntax.Node) (int, int) {
	cache := e.file.Memo(expressionSizesKey{}, func() any { return &sync.Map{} }).(*sync.Map)
	if cached, ok := cache.Load(scope); ok {
		sizes := cached.([2]int)
		return sizes[0], sizes[1]
	}
	values, before := e.countExpressions(scope)
	cache.Store(scope, [2]int{values, before})
	return values, before
}

func (e *Env) countExpressions(scope syntax.Node) (values, before int) {
	count := func(x syntax.Node) bool {
		switch x.(type) {
		case *syntax.Block, *syntax.Namespace, *syntax.Function, *syntax.Method, *syntax.ClassLike,
			*syntax.If, *syntax.ElseIf, *syntax.Else, *syntax.While, *syntax.DoWhile, *syntax.For,
			*syntax.Try, *syntax.Finally, *syntax.Switch, *syntax.Case, *syntax.Label, *syntax.Goto, *syntax.Break, *syntax.Continue:
		case syntax.Expr:
			switch n := x.(type) {
			case *syntax.Name, *syntax.Identifier, *syntax.ArgList, *syntax.VariadicPlaceholder, *syntax.ArgPlaceholder:
			case *syntax.Variable:
				assignment, target := n.Parent().(*syntax.Assign)
				if !target || assignment.Var != n || assignment.Op.Kind != syntax.TEqual || n.NameExpr != nil {
					values++
					before++
				}
			default:
				values++
				before++
			}

		case syntax.Stmt:
			before++
		}
		return !syntax.IsVariableScope(x)
	}
	if scope == nil {
		syntax.InspectFile(e.file, count)
	} else {
		syntax.Inspect(syntax.VariableScopeBody(scope), count)
	}
	return values, before
}
