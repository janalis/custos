// Package enumfromuncheckedexternalvalue implements the EnumFromUncheckedExternalValue inspection.
package enumfromuncheckedexternalvalue

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "EnumFromUncheckedExternalValue" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KStaticCall} }

const message = "Handle unknown enum backing values explicitly."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP81 {
		return
	}
	call := n.(*syntax.StaticCall)
	name, ok := call.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(name.Value, "from") {
		return
	}
	class := ctx.Index().Class(semanticquery.StaticCallClass(ctx, call.Class), ctx.PHP)
	if class == nil || class.Kind != syntax.KindEnum {
		return
	}
	if !ctx.Index().IsSubtype(class.FQN, "BackedEnum", ctx.PHP) {
		return
	}
	value := semanticquery.CallArgument(call.Args, 0, "value")
	fetch, ok := value.(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	source, ok := fetch.Var.(*syntax.Variable)
	if !ok || (source.Name != "_GET" && source.Name != "_POST" && source.Name != "_REQUEST" && source.Name != "_COOKIE") {
		return
	}
	// A closure or arrow function body runs later, outside any guard or try
	// block that lexically surrounds it, so the walk stops at that boundary.
	for p := call.Parent(); p != nil && !syntax.IsFuncLike(p); p = p.Parent() {
		if guard, ok := p.(*syntax.If); ok && guard.Body.Span().Contains(call.Span()) {
			stmts, listed := syntax.StmtListOf(guard.Body)
			if listed && len(stmts) > 0 && stmts[0].Span().Contains(call.Span()) && provesMembership(ctx, guard.Cond, value, class) {
				return
			}
		}
		// Only the try body is protected: a call in a catch or finally
		// block throws past this statement's own catch clauses.
		if tr, ok := p.(*syntax.Try); ok && tr.Body.Span().Contains(call.Span()) {
			for _, catch := range tr.Catches {
				for _, name := range catch.Types {
					fq := ctx.Names().Class(name.Value, name.Span().Start)
					if fq == "ValueError" || fq == "Error" || fq == "Throwable" {
						return
					}
				}
			}
		}
	}
	ctx.ReportNode(call, message)
}

// provesMembership reports whether cond ends with a strict in_array check of
// value against a non-empty literal list made only of declared backing
// values. Only the last operand of a conjunction qualifies: everything before
// it is evaluated first, so nothing runs between the check and the guarded use.
func provesMembership(ctx *analysis.Context, cond, value syntax.Expr, class *index.Class) bool {
	cond = syntax.UnwrapParens(cond)
	for {
		and, ok := cond.(*syntax.Binary)
		if !ok || (and.Op.Kind != syntax.TBooleanAnd && and.Op.Kind != syntax.TAnd) {
			break
		}
		cond = syntax.UnwrapParens(and.Right)
	}
	check, _ := semanticquery.GlobalCall(ctx, cond, "in_array")
	if check == nil || !astquery.Equivalent(ctx.File, semanticquery.CallArgument(check.Args, 0, "needle"), value) {
		return false
	}
	if strict, known := astquery.BoolConst(semanticquery.CallArgument(check.Args, 2, "strict")); !strict || !known {
		return false
	}
	allowed := semanticquery.NativeArray(ctx, semanticquery.CallArgument(check.Args, 1, "haystack"))
	if allowed == nil || len(allowed.Items) == 0 {
		return false
	}
	for _, item := range allowed.Items {
		if !isBackingValue(class, ctx.Text(item.Value)) {
			return false
		}
	}
	return true
}

// isBackingValue reports whether the literal source text names a declared
// case value: the same text, or a string literal decoding to the same bytes
// whatever its quote style.
func isBackingValue(class *index.Class, text string) bool {
	decoded, quoted := astquery.StringLiteralValue(text)
	for _, constant := range class.Consts {
		if !constant.Case {
			continue
		}
		if constant.Value == text {
			return true
		}
		if declared, ok := astquery.StringLiteralValue(constant.Value); ok && quoted && declared == decoded {
			return true
		}
	}
	return false
}

func (rule) Semantic() {}
