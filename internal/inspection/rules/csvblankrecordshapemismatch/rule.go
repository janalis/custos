// Package csvblankrecordshapemismatch implements the native CsvBlankRecordShapeMismatch inspection.
package csvblankrecordshapemismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Recognize a blank CSV record as a null field."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CsvBlankRecordShapeMismatch" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TIsIdentical && b.Op.Kind != syntax.TIsNotIdentical {
		return
	}
	var source syntax.Expr
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		if arr := semanticquery.NativeArray(ctx, pair[1]); arr != nil && len(arr.Items) == 0 {
			source = pair[0]
		}
	}
	if source == nil {
		return
	}
	call, ok := semanticquery.NativeValue(ctx, source).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, call, "fgetcsv") {
		return
	}
	h := semanticquery.CallArgument(call.Args, 0, "stream")
	open, ok := semanticquery.NativeValue(ctx, h).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, open, "fopen") {
		return
	}
	path, k := semanticquery.NativeString(ctx, semanticquery.CallArgument(open.Args, 0, "filename"))
	if !k || path != "php://memory" {
		return
	}
	blank := false
	rewound := false
	for _, n := range semanticquery.NativePriorCalls(ctx, call, h, "fwrite", "fputs", "rewind", "fseek", "fread", "fgets", "fgetcsv") {
		c := n.(*syntax.FuncCall)
		switch semanticquery.NativeBuiltinName(ctx, c) {
		case "fwrite", "fputs":
			text, k := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 1, "data"))
			blank = k && (text == "\n" || text == "\r\n")
			rewound = false
		case "rewind":
			rewound = true
		case "fseek":
			offset, k := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 1, "offset"))
			rewound = k && offset == 0
		default:
			return
		}
	}
	if blank && rewound {
		ctx.ReportNode(b, message)
	}
}
