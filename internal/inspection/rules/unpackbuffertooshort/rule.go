// Package unpackbuffertooshort implements the native UnpackBufferTooShort inspection.
package unpackbuffertooshort

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Provide enough bytes for the unpack format."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UnpackBufferTooShort" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "unpack") {
		return
	}
	if ctx.PHP < phpversion.PHP71 && c.Args != nil && len(c.Args.Args) > 2 {
		return
	}
	format, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "format"))
	if !ok || len(format) > 4096 {
		return
	}
	if (ctx.PHP < phpversion.PHP56 && strings.ContainsAny(format, "JP")) || (ctx.PHP < phpversion.PHP72 && strings.ContainsAny(format, "gGeE")) || (ctx.PHP < phpversion.PHP55 && strings.Contains(format, "Z")) {
		return
	}
	data, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 1, "string"))
	if !ok {
		return
	}
	offset := int64(0)
	if e := semanticquery.CallArgument(c.Args, 2, "offset"); e != nil {
		var known bool
		offset, known = semanticquery.NativeInt(ctx, e)
		if !known || offset < 0 {
			return
		}
	}
	if missing(format, int64(len(data))-offset) {
		ctx.ReportNode(c, message)
	}
}

func missing(format string, size int64) bool {
	cursor := int64(0)
	bad := false
	for _, part := range strings.Split(format, "/") {
		if part == "" {
			return false
		}
		code := part[0]
		i := 1
		for i < len(part) && part[i] >= '0' && part[i] <= '9' {
			i++
		}
		count := int64(1)
		if i > 1 {
			var err error
			count, err = strconv.ParseInt(part[1:i], 10, 64)
			if err != nil || count > 1000000 {
				return false
			}
		}
		star := i < len(part) && part[i] == '*'
		width := int64(0)
		switch code {
		case 'c', 'C', 'a', 'A', 'Z', 'x':
			width = 1
		case 'n', 'v':
			width = 2
		case 'N', 'V', 'g', 'G':
			width = 4
		case 'J', 'P', 'e', 'E':
			width = 8
		case 'h', 'H':
			if star {
				cursor = size
				continue
			}
			count = (count + 1) / 2
			width = 1
		case 'X':
			if star {
				return false
			}
			cursor -= count
			if cursor < 0 {
				return false
			}
			continue
		case '@':
			if star {
				return false
			}
			cursor = count
			bad = bad || cursor > size
			continue
		default:
			return false
		}
		if star {
			cursor = size
			continue
		}
		cursor += width * count
		bad = bad || cursor > size
	}
	return bad
}
