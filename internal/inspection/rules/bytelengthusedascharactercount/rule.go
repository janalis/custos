// Package bytelengthusedascharactercount implements the ByteLengthUsedAsCharacterCount inspection.
package bytelengthusedascharactercount

import (
	"unicode/utf8"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ByteLengthUsedAsCharacterCount" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

const message = "Count Unicode characters with the intended counting model."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call, _ := semanticquery.GlobalCall(ctx, n, "strlen")
	if call == nil {
		return
	}
	text, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "string"))
	if known && utf8.ValidString(text) && utf8.RuneCountInString(text) < len(text) && !feedsByteOperation(ctx, call) {
		ctx.ReportNode(call, message)
	}
}

// byteOperations are builtins whose offsets, lengths and sizes are counted in
// bytes: a strlen() result handed to one of them is a byte quantity on
// purpose.
var byteOperations = []string{
	"substr", "substr_replace", "substr_count", "substr_compare", "mb_strcut",
	"strpos", "stripos", "strrpos", "strripos", "strncmp", "strncasecmp",
	"str_pad", "pack", "fread", "fwrite", "fputs", "fgets", "fseek", "ftruncate",
	"stream_get_contents", "stream_copy_to_stream", "file_get_contents",
	"socket_read", "socket_write", "socket_send", "socket_recv",
}

// feedsByteOperation reports whether n, possibly parenthesised, negated or
// adjusted by integer arithmetic, is an argument of a builtin byte operation.
func feedsByteOperation(ctx *analysis.Context, n syntax.Node) bool {
	for {
		parent, child := astquery.ParentSkipParens(n)
		switch p := parent.(type) {
		case *syntax.Binary:
			if p.Op.Kind != syntax.TPlus && p.Op.Kind != syntax.TMinus && p.Op.Kind != syntax.TMul {
				return false
			}
		case *syntax.Unary:
			if p.Op.Kind != syntax.TMinus {
				return false
			}
		default:
			outer := astquery.ParentFuncCall(child)
			if outer == nil {
				return false
			}
			consumer, _ := semanticquery.GlobalCall(ctx, outer, byteOperations...)
			return consumer != nil
		}
		n = parent
	}
}

func (rule) Semantic() {}
