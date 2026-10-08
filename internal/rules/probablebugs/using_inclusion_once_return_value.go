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
	if inclusionOnceTestedForSuccess(inc) { // E3
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
	var n syntax.Node = inc
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
		break
	}
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
