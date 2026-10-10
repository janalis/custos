// Package regexbyteoffsetascharacteroffset implements the RegexByteOffsetAsCharacterOffset inspection.
package regexbyteoffsetascharacteroffset

import (
	"strings"
	"unicode/utf8"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RegexByteOffsetAsCharacterOffset" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

const message = "Convert byte offsets before using character-based slicing."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call, function := semanticquery.GlobalCall(ctx, n, "mb_substr", "grapheme_substr")
	if call == nil {
		return
	}
	if function == "mb_substr" {
		if encoding := semanticquery.CallArgument(call.Args, 3, "encoding"); encoding != nil {
			charset, known := semanticquery.NativeString(ctx, encoding)
			if !known || (!strings.EqualFold(charset, "UTF-8") && !strings.EqualFold(charset, "UTF8")) {
				return
			}
		}
	}
	subject := semanticquery.CallArgument(call.Args, 0, "string")
	text, known := semanticquery.NativeString(ctx, subject)
	if !known || !utf8.ValidString(text) || utf8.RuneCountInString(text) == len(text) {
		return
	}
	offset, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(call.Args, 1, "start")).(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	index, known := semanticquery.NativeInt(ctx, offset.Dim)
	if !known || index != 1 {
		return
	}
	match, ok := offset.Var.(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	variable, ok := match.Var.(*syntax.Variable)
	if !ok {
		return
	}
	prior := flowquery.NativePriorStatements(ctx.File, call)
	if len(prior) == 0 {
		return
	}
	statement, ok := prior[len(prior)-1].(*syntax.ExprStmt)
	if !ok {
		return
	}
	expression := statement.Expr
	if assign, ok := expression.(*syntax.Assign); ok {
		expression = assign.Value
	}
	preg, _ := semanticquery.GlobalCall(ctx, expression, "preg_match")
	if preg == nil {
		return
	}
	captured := semanticquery.CallArgument(preg.Args, 2, "matches")
	if !astquery.Equivalent(ctx.File, captured, variable) {
		return
	}
	source, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(preg.Args, 1, "subject"))
	if !known || source != text {
		return
	}
	has, known := semanticquery.NativeFlagContains(ctx, semanticquery.CallArgument(preg.Args, 3, "flags"), "PREG_OFFSET_CAPTURE")
	if known && has {
		ctx.ReportNode(call, message)
	}
}

func (rule) Semantic() {}
