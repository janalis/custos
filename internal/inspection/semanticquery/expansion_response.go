package semanticquery

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// ExpansionHeader reads a resolved literal header and its replacement policy.
func ExpansionHeader(ctx *analysis.Context, c *syntax.FuncCall) (key, value string, replace, known bool) {
	if !NativeBuiltin(ctx, c, "header") {
		return "", "", false, false
	}
	text, ok := NativeString(ctx, CallArgument(c.Args, 0, "header"))
	if !ok || strings.ContainsAny(text, "\r\n") {
		return "", "", false, false
	}
	k, v, ok := strings.Cut(text, ":")
	if !ok {
		return "", "", false, false
	}
	replace = true
	if e := CallArgument(c.Args, 1, "replace"); e != nil {
		replace, ok = NativeTruth(ctx, e)
		if !ok {
			return "", "", false, false
		}
	}
	return strings.ToLower(strings.TrimSpace(k)), strings.TrimSpace(v), replace, true
}

// ExpansionResponse describes proven header state and a complete response body.
// Previous fields concern the point immediately before the queried header.
type ExpansionResponse struct {
	PreviousLengths        []string
	PreviousVary           string
	HasLength, HasTransfer bool
	Final                  bool
	Status                 int64
	BodyLength             int64
	BodyKnown              bool
	Coding                 string
}

// ExpansionResponseAt caches one bounded scan of each direct statement list.
// Unknown operations invalidate state; output transforms invalidate body proof.
func ExpansionResponseAt(ctx *analysis.Context, c *syntax.FuncCall) ExpansionResponse {
	index := ctx.Memo("expansion.response.headers", func() any {
		result := map[*syntax.FuncCall]ExpansionResponse{}
		expansionResponseList(ctx, ctx.File.Stmts, true, result)
		syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
			switch p := n.(type) {
			case *syntax.Block:
				expansionResponseList(ctx, p.Stmts, false, result)
			case *syntax.Namespace:
				expansionResponseList(ctx, p.Stmts, true, result)
			}
			return true
		})
		return result
	}).(map[*syntax.FuncCall]ExpansionResponse)
	return index[c]
}

func expansionResponseList(ctx *analysis.Context, list []syntax.Stmt, top bool, out map[*syntax.FuncCall]ExpansionResponse) {
	if len(list) > 4096 {
		return
	}
	fields := map[string][]string{}
	final := map[string]*syntax.FuncCall{}
	var headers []*syntax.FuncCall
	body := int64(0)
	bodyKnown := true
	status := int64(200)
	complete := top
	coding := ""
	terminated := false
	for _, statement := range list {
		if terminated {
			break
		}
		switch st := statement.(type) {
		case *syntax.Echo:
			for _, value := range st.Exprs {
				n, ok := expansionByteLength(ctx, value, 0)
				if !ok || n > 1<<30-body {
					bodyKnown = false
				} else {
					body += n
				}
			}
		case *syntax.ExprStmt:
			switch value := syntax.UnwrapParens(st.Expr).(type) {
			case *syntax.Exit:
				complete = true
				terminated = true
				if value.Args != nil {
					n, ok := expansionByteLength(ctx, CallArgument(value.Args, 0, "status"), 0)
					bodyKnown = bodyKnown && ok
					if ok {
						body += n
					}
				}
			case *syntax.Print:
				n, ok := expansionByteLength(ctx, value.Expr, 0)
				bodyKnown = bodyKnown && ok
				if ok {
					body += n
				}
			case *syntax.FuncCall:
				switch NativeBuiltinName(ctx, value) {
				case "header":
					key, text, replace, ok := ExpansionHeader(ctx, value)
					if !ok {
						bodyKnown = false
						clear(fields)
						clear(final)
						continue
					}
					// Redirect and CGI status headers can change status implicitly.
					if key == "location" || key == "status" {
						bodyKnown = false
					}
					prev := ExpansionResponse{PreviousLengths: append([]string(nil), fields["content-length"]...), PreviousVary: strings.Join(fields["vary"], ",")}
					if replace {
						fields[key] = []string{text}
					} else {
						fields[key] = append(fields[key], text)
					}
					final[key] = value
					prev.HasLength = len(fields["content-length"]) > 0
					prev.HasTransfer = len(fields["transfer-encoding"]) > 0
					out[value] = prev
					headers = append(headers, value)
					if e := CallArgument(value.Args, 2, "response_code"); e != nil {
						v, ok := NativeInt(ctx, e)
						if !ok {
							bodyKnown = false
						} else if v != 0 {
							status = v
						}
					}
				case "header_remove":
					e := CallArgument(value.Args, 0, "name")
					if e == nil {
						clear(fields)
						clear(final)
					} else if name, ok := NativeString(ctx, e); ok {
						delete(fields, strings.ToLower(name))
						delete(final, strings.ToLower(name))
					} else {
						clear(fields)
						clear(final)
						bodyKnown = false
					}
				case "http_response_code":
					e := CallArgument(value.Args, 0, "response_code")
					if e != nil {
						v, ok := NativeInt(ctx, e)
						if !ok {
							bodyKnown = false
						} else {
							status = v
						}
					}
				default:
					clear(fields)
					clear(final)
					bodyKnown = false
				}
			default:
				clear(fields)
				clear(final)
				bodyKnown = false
			}
		default:
			clear(fields)
			clear(final)
			bodyKnown = false
		}
	}
	if len(fields["content-encoding"]) == 1 {
		coding = strings.ToLower(fields["content-encoding"][0])
	}
	allowed := complete && bodyKnown && status >= 200 && status != 204 && status != 304 && len(fields["transfer-encoding"]) == 0
	for _, header := range headers {
		record := out[header]
		key, _, _, _ := ExpansionHeader(ctx, header)
		record.Final = final[key] == header
		record.Status = status
		record.BodyLength = body
		record.BodyKnown = allowed
		record.Coding = coding
		out[header] = record
	}
}

func expansionByteLength(ctx *analysis.Context, e syntax.Expr, depth int) (int64, bool) {
	if depth > 8 {
		return 0, false
	}
	if text, ok := NativeString(ctx, e); ok {
		return int64(len(text)), true
	}
	switch value := NativeValue(ctx, e).(type) {
	case *syntax.Binary:
		if value.Op.Kind != syntax.TDot {
			return 0, false
		}
		a, ak := expansionByteLength(ctx, value.Left, depth+1)
		b, bk := expansionByteLength(ctx, value.Right, depth+1)
		return a + b, ak && bk && a+b <= 1<<30
	case *syntax.FuncCall:
		if !NativeBuiltin(ctx, value, "str_repeat") {
			return 0, false
		}
		n, known := NativeInt(ctx, CallArgument(value.Args, 1, "times"))
		if !known || n < 0 || n > 1<<30 {
			return 0, false
		}
		size, known := expansionByteLength(ctx, CallArgument(value.Args, 0, "string"), depth+1)
		if !known || (size != 0 && n > (1<<30)/size) {
			return 0, false
		}
		return size * n, true
	}
	return 0, false
}

// ExpansionDecimalLength parses unsigned decimal lengths without overflow.
func ExpansionDecimalLength(s string) (int64, bool) {
	if s == "" || strings.Trim(s, "0123456789") != "" {
		return 0, false
	}
	value, err := strconv.ParseInt(s, 10, 64)
	return value, err == nil
}
