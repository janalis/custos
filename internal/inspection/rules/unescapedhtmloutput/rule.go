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
	parts := htmlInputs(ctx, operands)
	if len(parts) == 0 || inertResponse(ctx, n) {
		return
	}
	for _, part := range parts {
		if tainted(ctx, part.expr, 0, part.context) && !requestRewritten(ctx, n, part.expr, part.context) {
			ctx.ReportNode(part.expr, message)
		}
	}
}

// emitted is a request-derived part of an output statement and the context
// it is emitted into.
type emitted struct {
	expr    syntax.Expr
	context flow.Context
}

// htmlInputs returns the request-derived parts of an output statement
// emitted into a proven text node or quoted data attribute. Every earlier byte
// must be known, or come from a part that cannot hold markup: a number or a
// boolean anywhere, text escaped for HTML inside a text node, or an earlier
// request-derived part, read as the text it should be. Any other dynamic part
// could open an attribute or a raw-text element, which ends the search.
func htmlInputs(ctx *analysis.Context, operands []syntax.Expr) []emitted {
	var prefix strings.Builder
	var found []emitted
	open := true
	// earlier counts the parts found in earlier operands: echo emits each
	// operand before it evaluates the next one.
	earlier := 0
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
			// Wrapping guarantees a supported quote form; escapes still
			// follow the parsed source version. Text that does not decode
			// says nothing of what follows.
			text, decoded := syntax.StringLiteralValueForVersion(`"`+x.Raw+`"`, syntax.SourceVersion(x))
			prefix.WriteString(text)
			open = open && decoded
			return true
		}
		if text, known := semanticquery.NativeString(ctx, part); known {
			prefix.WriteString(text)
			return true
		}
		// The whole operand is evaluated before any of it is emitted, so a
		// later part with effects could still change the response.
		if len(found) > earlier && flowquery.MayHaveSideEffects(part) {
			return false
		}
		var partContext flow.Context
		if open {
			partContext = htmlContext(prefix.String())
		}
		switch {
		case !open:
		case partContext != 0 && tainted(ctx, part, 0, partContext):
			found = append(found, emitted{part, partContext})
			// What follows is read as if the part were the text intended.
			prefix.WriteString("x")
		case partContext == flow.HTML && ctx.Flow().Matched(part):
			// A switch label the request data was matched with is text of
			// the program: the parts after it are still looked at.
		case ctx.TypeOf(part).OnlyOf("int", "float", "bool", "null", "void"):
			// A number or a boolean adds no markup, and nothing prints no text.
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
		earlier = len(found)
		if !collect(operand) {
			return found[:earlier]
		}
	}
	return found
}

// escaped reports a part whose value derives from tracked inputs and was
// escaped for HTML on every path: it adds no markup to the text node.
func escaped(ctx *analysis.Context, part syntax.Expr, context flow.Context) bool {
	v := ctx.Flow().Value(part)
	return v.Complete && len(v.Sources) > 0 && v.Safe.Covers(context)
}

// htmlContext identifies what text written after prefix must be safe for:
// a text node, the text of a script or style element, or the value of a
// classified attribute, quoted or not.
func htmlContext(prefix string) flow.Context {
	// Lowering only ASCII letters keeps every byte where it is.
	lower := asciiLower(prefix)
	for _, element := range []struct {
		name    string
		context flow.Context
	}{{"script", flow.HTMLScript}, {"style", flow.HTMLRaw}} {
		open, closed := strings.LastIndex(lower, "<"+element.name), strings.LastIndex(lower, "</"+element.name)
		switch {
		case open > closed && strings.Contains(lower[open:], ">"):
			// The element's text runs until its end tag, whatever it holds;
			// a script's text decides whether JSON may be written there.
			if element.context == flow.HTMLScript {
				return scriptContract(prefix[open+strings.IndexByte(lower[open:], '>')+1:])
			}
			return element.context
		case open >= 0:
			return 0
		}
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
	switch {
	case tag && closing:
		// An end tag holds no attribute.
	case quote != 0:
		eq := quoteAt - 1
		for eq >= 0 && htmlSpace(prefix[eq]) {
			eq--
		}
		if eq < 0 || prefix[eq] != '=' {
			return 0
		}
		name, named := attributeName(prefix, eq)
		contract, classified := attributeContract(name, prefix[quoteAt+1:])
		if !named || !classified {
			return 0
		}
		if quote == '\'' {
			return contract | flow.HTMLSingleQuoted
		}
		return contract | flow.HTMLDoubleQuoted
	case tag:
		eq, value, unquoted := unquotedValue(prefix)
		if !unquoted {
			return 0
		}
		name, named := attributeName(prefix, eq)
		contract, classified := attributeContract(name, value)
		if !named || !classified {
			return 0
		}
		return contract | flow.HTMLUnquoted
	case text:
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
