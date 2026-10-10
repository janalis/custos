// Package csvreadlengthsplitsknownrecord implements the native CsvReadLengthSplitsKnownRecord inspection.
package csvreadlengthsplitsknownrecord

import (
	"net/url"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Allow enough bytes for the complete CSV record."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CsvReadLengthSplitsKnownRecord" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "fgetcsv") {
		return
	}
	h := semanticquery.CallArgument(c.Args, 0, "stream")
	if !semanticquery.ExpansionDUnaliased(ctx, h, c) {
		return
	}
	length, k := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 1, "length"))
	if !k || length <= 0 {
		return
	}
	f, ok := semanticquery.NativeLocalValue(ctx, h).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, f, "fopen") {
		return
	}
	source, sk := semanticquery.NativeString(ctx, semanticquery.CallArgument(f.Args, 0, "filename"))
	mode, mk := semanticquery.NativeString(ctx, semanticquery.CallArgument(f.Args, 1, "mode"))
	if !sk || !mk || mode != "r" && mode != "rb" || !strings.HasPrefix(source, "data://text/plain,") {
		return
	}
	record, err := url.PathUnescape(strings.TrimPrefix(source, "data://text/plain,"))
	if err != nil {
		return
	}
	if strings.ContainsAny(record, "\"\r") {
		return
	}
	if newline := strings.IndexByte(record, '\n'); newline >= 0 {
		record = record[:newline]
	}
	for _, arg := range []struct {
		position           int
		name, defaultValue string
	}{{2, "separator", ","}, {3, "enclosure", "\""}, {4, "escape", "\\"}} {
		e := semanticquery.CallArgument(c.Args, arg.position, arg.name)
		if e != nil {
			v, known := semanticquery.NativeString(ctx, e)
			if !known || v != arg.defaultValue {
				return
			}
		}
	}
	if len(semanticquery.NativeStreamCalls(ctx, c, h, "fgetcsv", "fgets", "fread", "fgetc", "fscanf", "stream_get_contents", "fpassthru", "fwrite", "ftruncate", "fseek", "rewind", "fclose")) != 0 {
		return
	}
	if length < int64(len(record)) {
		ctx.ReportNode(c, message)
	}
}
