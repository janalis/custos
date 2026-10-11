package usinginclusiononcereturnvalue

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
)

// usingInclusionOnceReturnValue reports include_once/require_once whose
// result is used as a value.
type usingInclusionOnceReturnValue struct{}

const usingInclusionOnceReturnValueMsg = "Only the first include_once/require_once returns the file's value; later ones return true."

func (usingInclusionOnceReturnValue) ID() string { return "UsingInclusionOnceReturnValue" }
func (usingInclusionOnceReturnValue) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KInclude}
}

func (usingInclusionOnceReturnValue) Check(ctx *analysis.Context, n syntax.Node) {
	inc := n.(*syntax.Include)
	kw := strings.ToLower(ctx.SpanText(inc.Keyword.Span))                                 // keywords are case-insensitive
	if !strings.HasSuffix(kw, "_once") || inc.Expr == nil || inc.Expr.Span().Len() == 0 { // D1
		return
	}
	if es, ok := inc.Parent().(*syntax.ExprStmt); ok && es.Expr == syntax.Expr(inc) { // D2
		return
	}
	if inclusionOnceTestedForSuccess(inc) || inclusionOnceFlagVariable(ctx, inc) { // E3
		return
	}
	// No fix: a plain include/require re-runs the file, which redeclares
	// the classes and functions it defines (fatal).
	ctx.Report(syntax.Span{Start: inc.Keyword.Span.Start, End: inc.Expr.Span().End}, usingInclusionOnceReturnValueMsg)
}

// inclusionOnceTestedForSuccess reports whether the inclusion's result
// (through parentheses and `@`) is discarded or only tested for success: a condition,
// a logical operand, compared with false, or cast to bool. include_once returns false
// only when the file cannot be included, so such tests are reliable.
func inclusionOnceTestedForSuccess(inc *syntax.Include) bool {
	return testedForSuccess(inc)
}

// inclusionOnceFlagVariable reports whether the inclusion's result is
// stored in a plain local variable (`$found = @include_once $f;`) that is
// never read (in a function) or only tested for success (`if ($found) break;`,
// `if (!$res) die();`): the variable is a success flag, which stays reliable.
// A variable other code can reach (global, static, by-reference parameter
// or closure use, or read by name) is not a flag.
func inclusionOnceFlagVariable(ctx *analysis.Context, inc *syntax.Include) bool {
	n := skipAtParens(inc)
	as, ok := n.Parent().(*syntax.Assign)
	if !ok || as.Op.Kind != syntax.TEqual || as.Value != n || as.ByRef {
		return false
	}
	target, ok := as.Var.(*syntax.Variable)
	if !ok {
		return false
	}
	if target.Name == "" {
		return false
	}
	if _, stmt := as.Parent().(*syntax.ExprStmt); !stmt && !testedForSuccess(as) {
		return false
	}
	scope := syntax.EnclosingFuncLike(inc)
	if boundByReference(scope, target.Name) {
		return false
	}
	reads := 0
	for _, a := range flowquery.VarAccesses(ctx.File, scope, target.Name) {
		switch a.By.(type) {
		case *syntax.Global, *syntax.StaticVar: // the value outlives the call
			return false
		}
		if a.ElemWrite || a.Compound || !a.Write && !testedForSuccess(a.Var) {
			return false
		}
		if !a.Write {
			reads++
		}
	}
	byName, included := scopeReadsByName(ctx.File, scope, inc)
	if byName {
		return false
	}
	if reads > 0 {
		return true
	}
	// A variable never read here may still be read by a file included in
	// the same scope; at file scope, by the file that includes this one.
	return scope != nil && !included
}

// boundByReference reports whether name is a by-reference parameter or
// by-reference closure use of scope: a write to it is seen by the caller.
func boundByReference(scope syntax.Node, name string) bool {
	for _, p := range syntax.FuncLikeParams(scope) {
		if p.ByRef && p.Var != nil && p.Var.Name == name {
			return true
		}
	}
	if cl, ok := scope.(*syntax.Closure); ok {
		for _, u := range cl.Uses {
			if u.ByRef && u.Var != nil && u.Var.Name == name {
				return true
			}
		}
	}
	return false
}

// scopeReadsByName scans the code of scope (nil: the file's top-level
// code), nested arrow functions included. byName is true when a local
// variable can be read without being named: compact(),
// get_defined_vars(), a variable-variable or eval. included is true when
// the scope contains an inclusion other than self, whose file shares the
// scope's variables.
func scopeReadsByName(f *syntax.File, scope syntax.Node, self *syntax.Include) (byName, included bool) {
	visit := func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.ClassLike, *syntax.Closure:
			return false // separate scope
		case *syntax.Eval:
			byName = true
		case *syntax.Variable:
			byName = byName || x.NameExpr != nil
		case *syntax.Include:
			included = included || x != self
		case *syntax.FuncCall:
			switch strings.ToLower(astquery.CallLastName(x)) {
			case "compact", "get_defined_vars":
				byName = true
			}
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(f, visit)
		return byName, included
	}
	_, body := astquery.ScopeParts(scope)
	syntax.Inspect(body, visit)
	return byName, included
}

// skipAtParens climbs from n through enclosing parentheses and `@`.
func skipAtParens(n syntax.Node) syntax.Node {
	for {
		parent := n.Parent()
		if u, ok := parent.(*syntax.Unary); ok && u.Op.Kind == syntax.TAt {
			n = u
			continue
		}
		if _, ok := parent.(*syntax.Paren); ok {
			n = parent
			continue
		}
		return n
	}
}

// testedForSuccess reports whether the value of n (through parentheses
// and `@`) is discarded or only tested for truth.
func testedForSuccess(start syntax.Node) bool {
	n := skipAtParens(start)
	if astquery.IsLogicalOperand(n) {
		return true
	}
	switch p := n.Parent().(type) {
	case *syntax.ExprStmt: // `@include_once $f;`: the result is discarded
		return true
	case *syntax.Unary: // custos: `(bool) include_once $f` is the success flag
		return p.Op.Kind == syntax.TBoolCast
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual:
			other := p.Right
			if p.Right == n {
				other = p.Left
			}
			v, ok := astquery.BoolConst(syntax.UnwrapParens(other))
			return ok && !v
		case syntax.TXor:
			return true
		}
	}
	return false
}
