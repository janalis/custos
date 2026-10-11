// Package reflectioncompositetypeassumednamed implements the native ReflectionCompositeTypeAssumedNamed inspection.
package reflectioncompositetypeassumednamed

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Inspect the composite reflection type members."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ReflectionCompositeTypeAssumedNamed" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.MethodCall)
	name, ok := call.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(name.Value, "getName") || ctx.PHP < phpversion.PHP80 {
		return
	}
	source, ok := semanticquery.NativeValue(ctx, call.Var).(*syntax.MethodCall)
	if !ok {
		return
	}
	method, ok := source.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	typ := ""
	if strings.EqualFold(method.Value, "getReturnType") {
		_, typ, _ = semanticquery.NativeReflectionTarget(ctx, source.Var)
	} else if strings.EqualFold(method.Value, "getType") {
		if param, ok := semanticquery.NativeValue(ctx, source.Var).(*syntax.ArrayDimFetch); ok {
			list, ok := semanticquery.NativeValue(ctx, param.Var).(*syntax.MethodCall)
			if !ok {
				return
			}
			id, ok := list.Name.(*syntax.Identifier)
			if !ok || !strings.EqualFold(id.Value, "getParameters") {
				return
			}
			params, _, known := semanticquery.NativeReflectionTarget(ctx, list.Var)
			i, ik := semanticquery.NativeInt(ctx, param.Dim)
			if !known || !ik || i < 0 || i >= int64(len(params)) {
				return
			}
			typ = params[i].Type
		} else if prop := semanticquery.NativeConstruction(ctx, source.Var, "ReflectionProperty"); prop != nil {
			cl, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(prop.Args, 0, "class"))
			pn, ok2 := semanticquery.NativeString(ctx, semanticquery.CallArgument(prop.Args, 1, "property"))
			if !ok || !ok2 {
				return
			}
			p := ctx.Index().FindProperty(cl, pn, ctx.PHP)
			if p != nil {
				typ = p.Type
			}
		}
	}
	if !composite(typ, ctx.PHP) || guarded(ctx, call, typ) {
		return
	}
	ctx.ReportNode(call, message)
}

// composite reports whether a declared type reflects as a union or
// intersection type. A leading `?` and a `null` member only make a single
// named type nullable, which still reflects as ReflectionNamedType.
func composite(typ string, php phpversion.Version) bool {
	if strings.Contains(typ, "&") {
		return php >= phpversion.PHP81
	}
	members := 0
	for m := range strings.SplitSeq(strings.TrimPrefix(typ, "?"), "|") {
		if !strings.EqualFold(strings.TrimSpace(m), "null") {
			members++
		}
	}
	return members > 1
}

// guarded reports whether an instanceof test on the receiver dominates the
// call: an enclosing branch, ternary/match arm or short-circuit operand whose
// condition proves a named type, or an earlier sibling statement that leaves
// the scope (or asserts) otherwise.
func guarded(ctx *analysis.Context, call *syntax.MethodCall, typ string) bool {
	named := func(cond syntax.Expr, truth bool) bool { return proves(ctx, call.Var, typ, cond, truth) }
	before := func(list []syntax.Stmt, at syntax.Node) bool {
		for _, s := range list[:max(syntax.StmtIndex(list, at), 0)] {
			switch x := s.(type) {
			case *syntax.If:
				// A leaving main body proves its condition false whatever the later arms do.
				if syntax.Terminates(x.Body) && named(x.Cond, false) || len(x.ElseIfs) == 0 && x.Else != nil && syntax.Terminates(x.Else.Body) && named(x.Cond, true) {
					return true
				}
			case *syntax.ExprStmt:
				c, ok := x.Expr.(*syntax.FuncCall)
				if ok && semanticquery.NativeBuiltin(ctx, c, "assert") && named(semanticquery.CallArgument(c.Args, 0, "assertion"), true) {
					return true
				}
			}
		}
		return false
	}
	var child syntax.Node = call
	for p := call.Parent(); ; child, p = p, p.Parent() {
		if p == nil {
			return before(ctx.File.Stmts, child)
		}
		if syntax.IsVariableScope(p) {
			return false
		}
		ok := false
		switch x := p.(type) {
		case *syntax.If:
			if child == syntax.Node(x.Body) {
				ok = named(x.Cond, true)
			} else if child != syntax.Node(x.Cond) {
				ok = named(x.Cond, false)
				for _, ei := range x.ElseIfs {
					if child == syntax.Node(ei) {
						break
					}
					ok = ok || named(ei.Cond, false)
				}
			}
		case *syntax.ElseIf:
			ok = child == syntax.Node(x.Body) && named(x.Cond, true)
		case *syntax.Ternary:
			ok = named(x.Cond, child == syntax.Node(x.Then)) && child != syntax.Node(x.Cond)
		case *syntax.Binary:
			if and, or := conjunction(x), disjunction(x); and || or {
				ok = child == syntax.Node(x.Right) && named(x.Left, and)
			}
		case *syntax.MatchArm:
			m, _ := x.Parent().(*syntax.Match)
			subject, _ := syntax.UnwrapParens(m.Cond).(*syntax.ConstFetch)
			ok = child == syntax.Node(x.Body) && len(x.Conds) > 0 && subject != nil && strings.EqualFold(strings.TrimPrefix(subject.Name.Value, `\`), "true")
			for _, cond := range x.Conds {
				ok = ok && named(cond, true)
			}
		default:
			if list, isList := syntax.StmtListOf(p); isList {
				ok = before(list, child)
			}
		}
		if ok {
			return true
		}
	}
}

func conjunction(b *syntax.Binary) bool {
	return b.Op.Kind == syntax.TBooleanAnd || b.Op.Kind == syntax.TAnd
}

func disjunction(b *syntax.Binary) bool {
	return b.Op.Kind == syntax.TBooleanOr || b.Op.Kind == syntax.TOr
}

// proves reports whether cond evaluating to truth establishes that recv is a
// ReflectionNamedType: a positive test for that class, or a failed test for
// the composite class the declared type reflects as.
func proves(ctx *analysis.Context, recv syntax.Expr, typ string, cond syntax.Expr, truth bool) bool {
	switch x := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && proves(ctx, recv, typ, x.Expr, !truth)
	case *syntax.Binary:
		and := conjunction(x)
		if !and && !disjunction(x) {
			return false
		}
		l, r := proves(ctx, recv, typ, x.Left, truth), proves(ctx, recv, typ, x.Right, truth)
		if and == truth {
			return l || r
		}
		return l && r
	case *syntax.Instanceof:
		cl, ok := x.Class.(*syntax.Name)
		if !ok || !semanticquery.NativeSameValue(ctx, x.Expr, recv) {
			return false
		}
		want := "ReflectionNamedType"
		if !truth {
			want = "ReflectionUnionType"
			if !strings.Contains(typ, "|") {
				want = "ReflectionIntersectionType"
			}
		}
		return strings.EqualFold(ctx.Names().Class(cl.Value, cl.Span().Start), want)
	}
	return false
}
