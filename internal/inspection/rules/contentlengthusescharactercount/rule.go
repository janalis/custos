// Package contentlengthusescharactercount implements the native ContentLengthUsesCharacterCount inspection.
package contentlengthusescharactercount

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Measure response content length in bytes."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ContentLengthUsesCharacterCount" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	var fixes []diagnostic.Fix

	if name != "header" {
		return
	}
	value := semanticquery.CallArgument(c.Args, 0, "header")
	binary, ok := value.(*syntax.Binary)
	if !ok || binary.Op.Kind != syntax.TDot {
		return
	}
	prefix, known := semanticquery.NativeString(ctx, binary.Left)
	if !known || !strings.EqualFold(strings.TrimSpace(prefix), "Content-Length:") {
		return
	}
	length, ok := binary.Right.(*syntax.FuncCall)
	if !ok || (semanticquery.NativeBuiltinName(ctx, length) != "mb_strlen" && semanticquery.NativeBuiltinName(ctx, length) != "grapheme_strlen") {
		return
	}
	body := semanticquery.CallArgument(length.Args, 0, "string")
	if semanticquery.NativeBuiltinName(ctx, length) == "mb_strlen" && lengthCountsBytes(ctx, length) {
		return
	}
	if !emittedBody(ctx, c, body) {
		return
	}
	if _, bound := ctx.Flow().BoundArguments(length); bound && body != nil && lengthEncodingSafe(ctx, length) && semanticquery.ArgumentsRemovable(length) && semanticquery.CommentFreeSpan(ctx, length.Span()) {
		fixes = append(fixes, semanticquery.ReplacementFix(length.Span(), semanticquery.GlobalCallName(ctx, "strlen", c)+"("+ctx.Text(body)+")", "Count response bytes"))
	}

	ctx.ReportNode(c, message, fixes...)
}

func lengthEncodingSafe(ctx *analysis.Context, call *syntax.FuncCall) bool {
	encoding, state := semanticquery.CallArgumentBinding(call.Args, 1, "encoding")
	if state == semanticquery.ArgumentMissing {
		return true
	}
	name, known := semanticquery.NativeString(ctx, encoding)
	if !known {
		return false
	}
	switch strings.ToUpper(name) {
	case "UTF-8", "UTF8", "UTF-16", "UTF-16LE", "UTF-16BE", "UTF-32", "UTF-32LE", "UTF-32BE":
		return true
	}
	return false
}

func emittedBody(ctx *analysis.Context, header *syntax.FuncCall, body syntax.Expr) bool {
	if body == nil {
		return false
	}
	if text, known := semanticquery.NativeString(ctx, body); known {
		ascii := true
		for _, c := range []byte(text) {
			if c >= 128 {
				ascii = false
			}
		}
		if ascii {
			return false
		}
	}
	if _, known := semanticquery.NativeInt(ctx, body); known {
		return false
	}
	if literal, ok := semanticquery.NativeValue(ctx, body).(*syntax.Literal); ok && literal.LitKind == syntax.LitFloat {
		return false
	}
	if constant, ok := semanticquery.NativeValue(ctx, body).(*syntax.ConstFetch); ok {
		name := strings.ToLower(semanticquery.GlobalConstName(ctx, constant))
		if name == "true" || name == "false" || name == "null" {
			return false
		}
	}
	for _, call := range priorGlobal(ctx, header, "ob_start", "ini_set") {
		if semanticquery.NativeBuiltinName(ctx, call) == "ob_start" && !semanticquery.CallArgumentOmitted(call.Args, 0, "callback") {
			return false
		}
		if key, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "option")); known && key == "zlib.output_compression" {
			return false
		}
	}
	owner := syntax.EnclosingVariableScope(header)
	for _, call := range ctx.Flow().Calls(owner) {
		if call.Node.Span().Start > header.Span().End {
			return false
		}
	}
	outputs := 0
	matched := false
	syntax.InspectFile(ctx.File, func(node syntax.Node) bool {
		if syntax.EnclosingVariableScope(node) != owner || node.Span().Start < header.Span().End {
			return true
		}
		if _, ok := node.(*syntax.Print); ok {
			outputs++
			return true
		}
		// Template text between PHP tags is one more response chunk.
		if html, ok := node.(*syntax.InlineHTML); ok && html.Raw != "" {
			outputs++
			return true
		}
		echo, ok := node.(*syntax.Echo)
		if !ok {
			return true
		}
		outputs++
		if len(echo.Exprs) == 1 && semanticquery.NativeDominates(header, echo) && !repeatedOutput(echo, owner) && semanticquery.SameOperand(ctx, body, echo.Exprs[0]) {
			matched = true
		}
		return true
	})
	return matched && outputs == 1
}

// repeatedOutput reports whether a loop inside owner may run the output
// statement more than once.
func repeatedOutput(output, owner syntax.Node) bool {
	for p := output.Parent(); p != nil && p != owner; p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			return true
		}
	}
	return false
}

func priorGlobal(ctx *analysis.Context, at syntax.Node, names ...string) []*syntax.FuncCall {
	out := []*syntax.FuncCall{}
	for _, c := range ctx.Flow().Calls(syntax.EnclosingVariableScope(at)) {
		call, ok := c.Node.(*syntax.FuncCall)
		if !ok || !semanticquery.NativeDominates(call, at) {
			continue
		}
		for _, name := range names {
			if c.Name == name {
				out = append(out, call)
				break
			}
		}
	}
	return out
}

// lengthCountsBytes reports whether the resolved mb_strlen encoding maps
// every byte to one character, so the call already measures bytes.
func lengthCountsBytes(ctx *analysis.Context, call *syntax.FuncCall) bool {
	name, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 1, "encoding"))
	if !known {
		return false
	}
	switch strings.ToUpper(name) {
	case "8BIT", "BINARY", "ASCII", "US-ASCII", "LATIN1",
		"ISO-8859-1", "ISO-8859-2", "ISO-8859-3", "ISO-8859-4", "ISO-8859-5", "ISO-8859-6", "ISO-8859-7", "ISO-8859-8",
		"ISO-8859-9", "ISO-8859-10", "ISO-8859-13", "ISO-8859-14", "ISO-8859-15", "ISO-8859-16",
		"WINDOWS-1251", "WINDOWS-1252", "WINDOWS-1254", "CP1251", "CP1252", "CP1254":
		return true
	}
	return false
}
