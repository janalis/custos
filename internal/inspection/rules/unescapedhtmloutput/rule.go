// Package unescapedhtmloutput implements the native UnescapedHtmlOutput inspection.
package unescapedhtmloutput

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
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
	input, context := htmlInput(ctx, operands)
	if input == nil || !tainted(ctx, input, 0, context) || requestRewritten(ctx, n, input, context) || inertResponse(ctx, n) {
		return
	}
	ctx.ReportNode(n, message)
}

// htmlInput returns the first request-derived part of an output statement
// emitted into a proven text node or quoted data attribute. Every earlier byte
// must be known, or come from a part that cannot hold markup: a number or a
// boolean anywhere, text escaped for HTML inside a text node. Any other
// dynamic part could open an attribute or a raw-text element, which ends the
// search.
func htmlInput(ctx *analysis.Context, operands []syntax.Expr) (syntax.Expr, flow.Context) {
	var prefix strings.Builder
	var input syntax.Expr
	var context flow.Context
	open := true
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
				// Wrapping guarantees a supported quote form; escapes still
				// follow the parsed source version.
				text, _ := syntax.StringLiteralValueForVersion(`"`+x.Raw+`"`, syntax.SourceVersion(x))
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
		var partContext flow.Context
		if input == nil && open {
			partContext = htmlContext(prefix.String())
		}
		switch {
		case input != nil:
			// The whole operand is evaluated before any of it is emitted, so
			// a later part with effects could still change the response.
			return !flowquery.MayHaveSideEffects(part)
		case !open:
		case partContext != 0 && tainted(ctx, part, 0, partContext):
			input, context = part, partContext
		case partContext == flow.HTML && ctx.Flow().Matched(part):
			// A switch label the request data was matched with is text of
			// the program: the parts after it are still looked at.
		case ctx.TypeOf(part).OnlyOf("int", "float", "bool"):
		case partContext != 0 && escaped(ctx, part, partContext):
			// Escaping tracked inputs does not prove that an opaque value
			// contains no trusted quote delimiters of its own.
			open = partContext == flow.HTML
		default:
			open = false
		}
		return true
	}
	for _, operand := range operands {
		if !collect(operand) {
			return nil, 0
		}
		if input != nil {
			break
		}
	}
	return input, context
}

// escaped reports a part whose value derives from tracked inputs and was
// escaped for HTML on every path: it adds no markup to the text node.
func escaped(ctx *analysis.Context, part syntax.Expr, context flow.Context) bool {
	v := ctx.Flow().Value(part)
	return v.Complete && len(v.Sources) > 0 && v.Safe&context != 0
}

// htmlContext identifies text nodes and quoted ordinary data attributes.
func htmlContext(prefix string) flow.Context {
	lower := strings.ToLower(prefix)
	if strings.Contains(lower, "<script") || strings.Contains(lower, "<style") {
		return 0
	}
	tag, text, closing := false, false, false
	quoteAt := 0
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
				quote, quoteAt = c, i
			case '>':
				tag, text = false, true
			case '<':
				return 0
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
				return 0
			}
			tag, closing = true, start != i+1
		}
	}
	if quote != 0 && !closing && ordinaryAttribute(prefix, quoteAt) {
		if quote == '\'' {
			return flow.HTMLSingleQuoted
		}
		return flow.HTMLDoubleQuoted
	}
	if text && !tag {
		return flow.HTML
	}
	return 0
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
