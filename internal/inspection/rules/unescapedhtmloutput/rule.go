// Package unescapedhtmloutput implements the native UnescapedHtmlOutput inspection.
package unescapedhtmloutput

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Escape untrusted text for its HTML output context."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UnescapedHtmlOutput" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KEcho, syntax.KPrint} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	var operands []syntax.Expr
	switch x := n.(type) {
	case *syntax.Echo:
		operands = x.Exprs
	case *syntax.Print:
		operands = []syntax.Expr{x.Expr}
	}
	input := textNodeInput(ctx, operands)
	if input == nil || !tainted(ctx, input, 0) || requestRewritten(ctx, n, input) || inertResponse(ctx, n) {
		return
	}
	ctx.ReportNode(n, message)
}

// textNodeInput returns the first dynamic part of an output statement when
// every byte emitted before it is known and leaves an HTML text node open.
// Earlier dynamic output could open an attribute or raw-text element.
func textNodeInput(ctx *analysis.Context, operands []syntax.Expr) syntax.Expr {
	var prefix strings.Builder
	var input syntax.Expr
	var collect func(syntax.Expr) bool
	collect = func(part syntax.Expr) bool {
		part = syntax.UnwrapParens(part)
		switch x := part.(type) {
		case *syntax.Binary:
			if x.Op.Kind == syntax.TDot {
				return collect(x.Left) && collect(x.Right)
			}
		case *syntax.InterpolatedString:
			if !x.Backtick {
				for _, piece := range x.Parts {
					if !collect(piece) {
						return false
					}
				}
				return true
			}
		case *syntax.StringPart:
			if input == nil {
				text, _ := astquery.StringLiteralValue(`"` + x.Raw + `"`)
				prefix.WriteString(text)
			}
			return true
		}
		if text, known := semanticquery.NativeString(ctx, part); known {
			if input == nil {
				prefix.WriteString(text)
			}
			return true
		}
		if input != nil {
			// The whole operand is evaluated before any of it is emitted, so
			// a later part with effects could still change the response.
			return !flowquery.MayHaveSideEffects(part)
		}
		input = part
		return true
	}
	for _, operand := range operands {
		if !collect(operand) {
			return nil
		}
		if input != nil {
			break
		}
	}
	if !textPrefix(prefix.String()) {
		return nil
	}
	return input
}

// textPrefix proves that known markup ends inside an ordinary text node.
func textPrefix(prefix string) bool {
	lower := strings.ToLower(prefix)
	if strings.Contains(lower, "<script") || strings.Contains(lower, "<style") {
		return false
	}
	tag, text := false, false
	var quote byte
	for i := 0; i < len(prefix); i++ {
		c := prefix[i]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		if tag {
			switch c {
			case '\'', '"':
				quote = c
			case '>':
				tag, text = false, true
			case '<':
				return false
			}
			continue
		}
		if c == '<' {
			start := i + 1
			if start < len(prefix) && prefix[start] == '/' {
				start++
			}
			// Comments, declarations and incomplete/unknown tags cannot
			// establish text context. Only ordinary ASCII tag names do.
			if start >= len(prefix) || lower[start] < 'a' || lower[start] > 'z' {
				return false
			}
			tag = true
		}
	}
	return text && !tag
}

// inertResponse reports whether the latest dominating literal Content-Type
// header selects a media type that browsers never parse as markup. Calls are
// listed in evaluation order, so a later header replaces an earlier one.
func inertResponse(ctx *analysis.Context, at syntax.Node) bool {
	inert := false
	for _, c := range ctx.Flow().Calls(syntax.EnclosingVariableScope(at)) {
		call, ok := c.Node.(*syntax.FuncCall)
		if !ok || c.Name != "header" || !semanticquery.NativeDominates(call, at) {
			continue
		}
		value, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "header"))
		name, media, _ := strings.Cut(strings.ToLower(value), ":")
		if !known || strings.TrimSpace(name) != "content-type" {
			continue
		}
		media, _, _ = strings.Cut(media, ";")
		media = strings.TrimSpace(media)
		inert = media == "text/plain" || media == "text/csv" || media == "application/json" || media == "application/octet-stream" || strings.HasSuffix(media, "+json")
	}
	return inert
}
