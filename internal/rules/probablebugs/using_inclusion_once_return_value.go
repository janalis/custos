package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// usingInclusionOnceReturnValue reports include_once/require_once whose
// result is used as a value.
type usingInclusionOnceReturnValue struct{}

func init() { register(usingInclusionOnceReturnValue{}) }

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
// stored in a local variable (`$found = @include_once $f;`) that is never
// read (in a function) or only tested for success (`if ($found) break;`, `if (!$res)
// die();`): the variable is a success flag, which stays reliable.
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
	reads := 0
	for _, a := range util.VarAccesses(ctx.File, scope, target.Name) {
		if a.ElemWrite || a.Compound || !a.Write && !testedForSuccess(a.Var) {
			return false
		}
		if !a.Write {
			reads++
		}
	}
	// A file-scope variable never read here may be read by the file
	// that includes this one.
	return reads > 0 || scope != nil
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
	if util.IsLogicalOperand(n) {
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
			v, ok := util.BoolConst(syntax.UnwrapParens(other))
			return ok && !v
		case syntax.TXor:
			return true
		}
	}
	return false
}
