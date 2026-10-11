// Package aeaddecryptiontaglengthunchecked implements the native AeadDecryptionTagLengthUnchecked inspection.
package aeaddecryptiontaglengthunchecked

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Validate the authentication tag length before decrypting."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "AeadDecryptionTagLengthUnchecked" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP71 {
		return
	}
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "openssl_decrypt") {
		return
	}
	cipher, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 1, "cipher_algo"))
	if !known {
		return
	}
	switch strings.ToLower(cipher) {
	case "aes-128-gcm", "aes-192-gcm", "aes-256-gcm":
	default:
		return
	}
	tag := semanticquery.CallArgument(call.Args, 5, "tag")
	if _, _, ok := request(ctx, tag); !ok {
		return
	}
	want := int64(ctx.Int("tagLength"))
	size, known := ctx.Flow().ExactLength(tag)
	if (!known || size != want) && !semanticquery.NativeLengthGuard(ctx, tag, want) && !lengthGuard(ctx, tag, want) {
		ctx.ReportNode(tag, message)
	}
}

// lengthGuard recognises the equivalent spellings of the exact-length proof
// that the shared guard leaves unknown: swapped operands, a ternary condition,
// and a guard reading the request element the tag comes from.
func lengthGuard(ctx *analysis.Context, tag syntax.Expr, want int64) bool {
	looped := false
	for p := tag.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch x := p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			looped = true
		case *syntax.If:
			if x.Body != nil && x.Body.Span().Contains(tag.Span()) && lengthCheck(ctx, x.Cond, true, tag, want, looped) {
				return true
			}
			if x.Else != nil && x.Else.Body.Span().Contains(tag.Span()) && lengthCheck(ctx, x.Cond, false, tag, want, looped) {
				return true
			}
		case *syntax.Ternary:
			if x.Then != nil && x.Then.Span().Contains(tag.Span()) && lengthCheck(ctx, x.Cond, true, tag, want, looped) {
				return true
			}
			if x.Else != nil && x.Else.Span().Contains(tag.Span()) && lengthCheck(ctx, x.Cond, false, tag, want, looped) {
				return true
			}
		}
		if st, ok := p.(syntax.Stmt); ok {
			// Any earlier guard of the same statement list dominates the
			// tag, as long as the tag variable is not used in between.
			list, i, _ := astquery.StmtList(ctx.File, st)
			for j := i - 1; j >= 0 && unused(list[j+1:i+1], tag); j-- {
				if guard, ok := list[j].(*syntax.If); ok && guard.Else == nil && len(guard.ElseIfs) == 0 && syntax.Terminates(guard.Body) && lengthCheck(ctx, guard.Cond, false, tag, want, looped) {
					return true
				}
			}
		}
	}
	return false
}

// unused reports whether no statement of between mentions the tag variable
// before the tag itself: a compound assignment or a by-reference argument
// would change the measured value without the flow facts noticing. A tag
// read straight from the request has no variable to watch.
func unused(between []syntax.Stmt, tag syntax.Expr) bool {
	name, ok := astquery.PlainVariableName(tag)
	free := true
	for _, st := range between {
		syntax.Inspect(st, func(n syntax.Node) bool {
			if v, isVar := n.(*syntax.Variable); ok && isVar && v.Name == name && v.Span().End <= tag.Span().Start {
				free = false
			}
			return free
		})
	}
	return free
}

// lengthCheck accepts a strict strlen comparison in either operand order that
// holds with the given truth and measures the value passed as the tag.
func lengthCheck(ctx *analysis.Context, condition syntax.Expr, truth bool, tag syntax.Expr, want int64, looped bool) bool {
	b, ok := syntax.UnwrapParens(condition).(*syntax.Binary)
	if !ok || !((truth && b.Op.Kind == syntax.TIsIdentical) || (!truth && b.Op.Kind == syntax.TIsNotIdentical)) {
		return false
	}
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		call, ok := syntax.UnwrapParens(pair[0]).(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, call, "strlen") {
			continue
		}
		n, known := semanticquery.NativeInt(ctx, pair[1])
		if known && n == want && sameTag(ctx, semanticquery.CallArgument(call.Args, 0, "string"), tag, looped) {
			return true
		}
	}
	return false
}

// sameTag ties the measured expression to the tag. A variable must still hold
// the value that was measured. Two reads of one unwritten request element agree
// too, unless a loop separates them: a later write in the loop body would reach
// the tag on the next iteration without passing the guard again.
func sameTag(ctx *analysis.Context, measured, tag syntax.Expr, looped bool) bool {
	if astquery.Equivalent(ctx.File, measured, tag) {
		mv, tv := ctx.Flow().Value(measured), ctx.Flow().Value(tag)
		if mv.Expr == tv.Expr && mv.Invalidated == tv.Invalidated {
			return true
		}
	}
	if looped {
		return false
	}
	global, key, ok := request(ctx, measured)
	if !ok {
		return false
	}
	tagGlobal, tagKey, _ := request(ctx, tag)
	return global == tagGlobal && key == tagKey
}

func request(ctx *analysis.Context, e syntax.Expr) (global, key string, ok bool) {
	x, ok := semanticquery.NativeValue(ctx, e).(*syntax.ArrayDimFetch)
	if !ok {
		return "", "", false
	}
	v, ok := x.Var.(*syntax.Variable)
	if !ok {
		return "", "", false
	}
	switch v.Name {
	case "_GET", "_POST", "_REQUEST", "_COOKIE", "_SERVER":
	default:
		return "", "", false
	}
	if !semanticquery.NativeRequestUnwritten(ctx, x) {
		return "", "", false
	}
	key, ok = semanticquery.NativeString(ctx, x.Dim)
	return v.Name, key, ok
}
