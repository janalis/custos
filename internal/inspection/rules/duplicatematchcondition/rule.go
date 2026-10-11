// Package duplicatematchcondition implements the DuplicateMatchCondition inspection.
package duplicatematchcondition

import (
	"fmt"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DuplicateMatchCondition" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMatch} }

const message = "Remove the repeated match condition."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP80 {
		return
	}
	seen := map[string]bool{}
	match := n.(*syntax.Match)
	for index, arm := range match.Arms {
		// The arm is dead only when every condition repeats one of an earlier
		// arm: seen is not updated until the keys of this arm are all known.
		keys := make([]string, len(arm.Conds))
		dead := len(arm.Conds) > 0
		for i, condition := range arm.Conds {
			keys[i] = conditionKey(ctx, condition)
			dead = dead && seen[keys[i]]
		}
		for i, condition := range arm.Conds {
			key := keys[i]
			if key == "" {
				continue
			}
			if seen[key] {
				fix, ok := diagnostic.Fix{}, false
				if dead {
					// One removal per arm, carried by its first condition.
					dead = false
					fix, ok = armRemoval(ctx, match, index)
				}
				if ok {
					ctx.ReportNode(condition, message, fix)
				} else {
					ctx.ReportNode(condition, message)
				}
			}
			seen[key] = true
		}
	}
}

// conditionKey returns a comparison key for a condition whose value is
// statically known, or "" when nothing definite can be said about it.
func conditionKey(ctx *analysis.Context, condition syntax.Expr) string {
	if v, ok := syntax.UnwrapParens(condition).(*syntax.Variable); ok && !stableLocal(ctx, v) {
		return ""
	}
	value := semanticquery.NativeValue(ctx, condition)
	if fetch, ok := condition.(*syntax.ClassConstFetch); ok {
		value = fetch
	}
	switch value := value.(type) {
	case *syntax.Literal:
		if value.LitKind == syntax.LitString {
			if text, ok := semanticquery.NativeString(ctx, value); ok {
				return "string:" + text
			}
			return ""
		}
		return fmt.Sprintf("%d:%s", value.LitKind, value.Raw)
	case *syntax.ClassConstFetch:
		name, ok := value.Name.(*syntax.Identifier)
		class := semanticquery.StaticCallClass(ctx, value.Class)
		if ok && class != "" && ctx.Index().FindConst(class, name.Value, ctx.PHP) != nil {
			return "case:" + class + "::" + name.Value
		}
	case *syntax.ConstFetch:
		name := strings.ToLower(value.Name.Value)
		if name == "true" || name == "false" || name == "null" {
			return name
		}
	}
	return ""
}

// stableLocal reports whether v is a local of a braced function body that is
// bound by exactly one plain assignment of a variable-free expression and
// cannot be rebound behind the flow resolver's back: no ++/--, compound,
// by-reference or destructuring
// write, no catch/foreach/global/static/unset binding, no by-reference
// closure capture, and no construct writing variables by name in the scope.
func stableLocal(ctx *analysis.Context, v *syntax.Variable) bool {
	scope := syntax.EnclosingFuncLike(v)
	if scope == nil {
		return false
	}
	body := syntax.FuncLikeBody(scope)
	if body == nil || v.NameExpr != nil {
		return false
	}
	writes := 0
	for _, access := range flowquery.VarAccesses(ctx.File, scope, v.Name) {
		if use, ok := access.By.(*syntax.ClosureUse); ok && use.ByRef {
			return false
		}
		if !access.Write {
			continue
		}
		assign, ok := access.By.(*syntax.Assign)
		if !ok || assign.Var != syntax.Expr(access.Var) || assign.ByRef || access.Compound {
			return false
		}
		writes++
		// The flow resolver follows the assigned expression into other
		// variables without this proof, so the value must not read any.
		syntax.Inspect(assign.Value, func(x syntax.Node) bool {
			if _, reads := x.(*syntax.Variable); reads {
				writes++
			}
			return true
		})
	}
	if writes != 1 {
		return false
	}
	dynamic := false
	syntax.Inspect(body, func(x syntax.Node) bool {
		switch x := x.(type) {
		case *syntax.Variable:
			dynamic = dynamic || x.NameExpr != nil
		case *syntax.Include, *syntax.Eval:
			dynamic = true
		case *syntax.FuncCall:
			dynamic = dynamic || strings.EqualFold(astquery.CallLastName(x), "extract")
		}
		return !dynamic
	})
	return !dynamic
}

// armRemoval builds the edit deleting arm index together with one adjacent
// comma; it is withheld when the deleted text may hold a comment.
func armRemoval(ctx *analysis.Context, match *syntax.Match, index int) (diagnostic.Fix, bool) {
	removal := match.Arms[index].Span()
	if next, ok := astquery.NextSignificant(ctx.File, removal.End); ok && next.Kind == syntax.TComma {
		removal.End = next.End
	} else if index > 0 {
		if previous, ok := astquery.NextSignificant(ctx.File, match.Arms[index-1].Span().End); ok && previous.Kind == syntax.TComma {
			removal.Start = previous.Start
		}
	}
	text := ctx.SpanText(removal)
	if strings.Contains(text, "/*") || strings.Contains(text, "//") || strings.Contains(text, "#") {
		return diagnostic.Fix{}, false
	}
	fix := astquery.ReplaceFix(removal, "")
	fix.Title = "Remove the unreachable match arm"
	return fix, true
}

func (rule) Semantic() {}
