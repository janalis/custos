package simplexmlloadfileusage

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// simpleXMLLoadFileUsage reports simplexml_load_file() calls (PHP bug #62577)
// and rewrites them to simplexml_load_string(file_get_contents(...)).
type simpleXMLLoadFileUsage struct{}

const simpleXMLLoadFileUsageMsg = "simplexml_load_file() is affected by PHP bug #62577; load the contents with file_get_contents() and parse them with simplexml_load_string()."

// Semantic marks the rule as needing the project index (user functions with
// the same names may be declared in other files).
func (simpleXMLLoadFileUsage) Semantic()  {}
func (simpleXMLLoadFileUsage) ID() string { return "SimpleXmlLoadFileUsage" }
func (simpleXMLLoadFileUsage) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (simpleXMLLoadFileUsage) Check(ctx *analysis.Context, n syntax.Node) {
	// custos: the bug needs the entity loader disabled with
	// libxml_disable_entity_loader(), deprecated from PHP 8.0 (libxml 2.9
	// disables external entities by default).
	if ctx.PHP >= phpversion.PHP80 {
		return
	}
	call := n.(*syntax.FuncCall)
	if !strings.EqualFold(astquery.CallLastName(call), "simplexml_load_file") || call.Args == nil || len(call.Args.Args) == 0 { // D1, D2/E1
		return
	}
	if !semanticquery.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, "simplexml_load_file") { // D1
		return
	}
	// Divergence: skip spreads, named arguments and `f(...)`.
	args := make([]*syntax.Arg, 0, len(call.Args.Args))
	for _, a := range call.Args.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok || arg.Unpack || arg.Name != nil {
			return
		}
		args = append(args, arg)
	}
	// F2: parsing a string drops the document's base URI, so the rewrite is
	// offered only when the options cannot make libxml load a relative
	// external resource.
	if len(args) > 2 && !optionsKeepBaseURIUnused(ctx, args[2].Value) {
		ctx.ReportNode(call, simpleXMLLoadFileUsageMsg)
		return
	}
	ctx.ReportNode(call, simpleXMLLoadFileUsageMsg, diagnostic.Fix{
		Title: "Use simplexml_load_string(file_get_contents(...))",
		Edits: func() []diagnostic.TextEdit {
			var b strings.Builder
			b.WriteString(semanticquery.QualifiedBuiltin(ctx, "simplexml_load_string", call.Span().Start))
			b.WriteByte('(')
			b.WriteString(semanticquery.QualifiedBuiltin(ctx, "file_get_contents", call.Span().Start))
			b.WriteByte('(')
			b.WriteString(ctx.Text(args[0]))
			b.WriteByte(')')
			for _, a := range args[1:] {
				b.WriteString(", ")
				b.WriteString(ctx.Text(a))
			}
			b.WriteByte(')')
			return []diagnostic.TextEdit{{Span: call.Span(), NewText: b.String()}}
		},
	})
}

// baseURIOptionBits are the libxml parser options that make the parser fetch
// resources named relative to the document: LIBXML_NOENT (2), LIBXML_DTDLOAD
// (4), LIBXML_DTDATTR (8), LIBXML_DTDVALID (16) and LIBXML_XINCLUDE (1024).
const baseURIOptionBits = 2 | 4 | 8 | 16 | 1024

// baseURIFreeOptions lists the libxml option constants that never make the
// parser resolve a relative system identifier.
var baseURIFreeOptions = map[string]bool{
	"LIBXML_BIGLINES":   true,
	"LIBXML_COMPACT":    true,
	"LIBXML_NOBLANKS":   true,
	"LIBXML_NOCDATA":    true,
	"LIBXML_NOEMPTYTAG": true,
	"LIBXML_NOERROR":    true,
	"LIBXML_NONET":      true,
	"LIBXML_NOWARNING":  true,
	"LIBXML_NOXMLDECL":  true,
	"LIBXML_NSCLEAN":    true,
	"LIBXML_PARSEHUGE":  true,
	"LIBXML_PEDANTIC":   true,
	"LIBXML_RECOVER":    true,
}

// optionsKeepBaseURIUnused reports whether the options expression is a
// constant provably free of every base-URI-dependent flag: an integer
// literal without those bits, a listed global LIBXML_* constant, or a `|`
// combination of such operands.
func optionsKeepBaseURIUnused(ctx *analysis.Context, e syntax.Expr) bool {
	switch v := e.(type) {
	case *syntax.Paren:
		return optionsKeepBaseURIUnused(ctx, v.Expr)
	case *syntax.Binary:
		return v.Op.Kind == syntax.TBar && optionsKeepBaseURIUnused(ctx, v.Left) && optionsKeepBaseURIUnused(ctx, v.Right)
	case *syntax.Literal:
		bits, ok := astquery.ParseIntLiteral(v.Raw)
		return ok && v.LitKind == syntax.LitInt && bits&baseURIOptionBits == 0
	case *syntax.ConstFetch:
		return baseURIFreeOptions[semanticquery.GlobalConstName(ctx, v)]
	}
	return false
}
