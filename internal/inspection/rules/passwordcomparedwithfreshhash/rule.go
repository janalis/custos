// Package passwordcomparedwithfreshhash implements the native PasswordComparedWithFreshHash inspection.
package passwordcomparedwithfreshhash

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Verify the password against the stored hash."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PasswordComparedWithFreshHash" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	// A comparison with false, null, a number, an array or the empty string
	// tests whether hashing succeeded; no stored hash takes part in it.
	if binary := n.(*syntax.Binary); notStoredHash(ctx, binary.Left) || notStoredHash(ctx, binary.Right) {
		return
	}
	check(ctx, n, message)
}

// notStoredHash reports whether e is known to hold a value that cannot be a
// stored password hash: a boolean, null, a number, an array or "".
func notStoredHash(ctx *analysis.Context, e syntax.Expr) bool {
	switch v := semanticquery.NativeValue(ctx, e).(type) {
	case *syntax.ConstFetch:
		switch strings.ToLower(strings.TrimPrefix(v.Name.Value, `\`)) {
		case "false", "null", "true":
			return true
		}
	case *syntax.Literal:
		if v.LitKind != syntax.LitString {
			return true
		}
		text, known := semanticquery.NativeString(ctx, v)
		return known && text == ""
	case *syntax.Array:
		return true
	}
	return false
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	binary := n.(*syntax.Binary)
	if ctx.PHP < phpversion.PHP55 {
		return
	}
	if binary.Op.Kind != syntax.TIsIdentical && binary.Op.Kind != syntax.TIsEqual && binary.Op.Kind != syntax.TIsNotIdentical && binary.Op.Kind != syntax.TIsNotEqual {
		return
	}
	for i, expr := range []syntax.Expr{binary.Left, binary.Right} {
		call, ok := semanticquery.NativeValue(ctx, expr).(*syntax.FuncCall)
		if !ok || semanticquery.NativeBuiltinName(ctx, call) != "password_hash" {
			continue
		}
		password := semanticquery.CallArgument(call.Args, 0, "password")
		stored := binary.Right
		if i == 1 {
			stored = binary.Left
		}
		if password == nil {
			return
		}
		var fixes []diagnostic.Fix
		fixable := syntax.UnwrapParens(expr) == call && !flowquery.MayHaveSideEffects(password) && !flowquery.MayHaveSideEffects(stored)
		_, bound := ctx.Flow().BoundArguments(call)
		algorithm, constant := semanticquery.NativeValue(ctx, semanticquery.CallArgument(call.Args, 1, "algo")).(*syntax.ConstFetch)
		fixable = fixable && bound && constant && (semanticquery.GlobalConstName(ctx, algorithm) == "PASSWORD_DEFAULT" || semanticquery.GlobalConstName(ctx, algorithm) == "PASSWORD_BCRYPT")
		if options := semanticquery.CallArgument(call.Args, 2, "options"); options != nil {
			entries, known := semanticquery.NativeArrayEntries(ctx, options)
			fixable = fixable && known && len(entries) == 0
		}
		if call.Args != nil {
			for _, node := range call.Args.Args {
				arg, ok := node.(*syntax.Arg)
				if !ok || arg.Unpack || arg.ByRef || flowquery.MayHaveSideEffects(arg.Value) {
					fixable = false
				}
			}
		}
		if fixable && binary.Op.Kind == syntax.TIsIdentical && semanticquery.CommentFreeSpan(ctx, binary.Span()) {
			fixes = append(fixes, semanticquery.ReplacementFix(binary.Span(), semanticquery.GlobalCallName(ctx, "password_verify", binary)+"("+ctx.Text(password)+", "+ctx.Text(stored)+")", "Verify the stored password hash"))
		}
		ctx.ReportNode(binary, message, fixes...)
		return
	}
}
