package curlsslserverspoofing

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// curlSslServerSpoofing reports CURLOPT_SSL_VERIFYHOST / CURLOPT_SSL_VERIFYPEER
// set to a value that disables TLS verification.
type curlSslServerSpoofing struct{}

const (
	curlHostMsg = "Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2."
	curlPeerMsg = "Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled."
)

func (curlSslServerSpoofing) ID() string { return "CurlSslServerSpoofing" }
func (curlSslServerSpoofing) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KConstFetch}
}

func (curlSslServerSpoofing) Check(ctx *analysis.Context, n syntax.Node) {
	k := n.(*syntax.ConstFetch) // the parser always gives it a name token
	var host bool
	switch astquery.LastNamePart(k.Name.Value) {
	case "CURLOPT_SSL_VERIFYHOST":
		host = true
	case "CURLOPT_SSL_VERIFYPEER":
	default:
		return // E5
	}
	site, value := curlSettingSite(ctx, k)
	if site == nil || value == nil {
		return
	}
	enable, disable := false, false
	for _, v := range semanticquery.DiscoverValues(ctx.Types(), value) {
		switch curlClassify(ctx, v, host) {
		case 1:
			enable = true
		case -1:
			disable = true
		}
	}
	if !disable || enable { // D6, E3
		return
	}
	if host {
		ctx.ReportNode(site, curlHostMsg)
	} else {
		ctx.ReportNode(site, curlPeerMsg)
	}
}

// curlSettingSite implements D1–D3: the reported node and the value set.
func curlSettingSite(ctx *analysis.Context, k *syntax.ConstFetch) (site syntax.Node, value syntax.Expr) {
	switch p := k.Parent().(type) {
	case *syntax.Arg: // D1 (an Arg's only expression child is its value)
		call := astquery.ParentFuncCall(k)
		if call == nil || !ctx.IsGlobalFunctionCall(call, "curl_setopt") || astquery.ArgCount(call) != 3 {
			return nil, nil
		}
		third, ok := call.Args.Args[2].(*syntax.Arg)
		if !ok || third.Value == nil {
			return nil, nil
		}
		return call, third.Value
	case *syntax.ArrayItem: // D2
		if p.Key != syntax.Expr(k) || p.Value == nil {
			return nil, nil
		}
		arr, ok := p.Parent().(*syntax.Array)
		if !ok || curlDestructuring(arr) { // list() / destructuring read the option
			return nil, nil
		}
		return p, p.Value
	case *syntax.ArrayDimFetch: // D3
		if p.Dim != syntax.Expr(k) {
			return nil, nil
		}
		var cur syntax.Expr = p
		for {
			up, ok := cur.Parent().(*syntax.ArrayDimFetch)
			if !ok || up.Var != cur {
				break
			}
			cur = up
		}
		a, ok := cur.Parent().(*syntax.Assign)
		if !ok || a.Var != cur || a.Op.Kind != syntax.TEqual || a.Value == nil {
			return nil, nil
		}
		return a, a.Value
	}
	return nil, nil
}

// curlDestructuring reports whether arr is (part of) a destructuring
// pattern: an assignment target or a foreach value, possibly nested in
// other arrays or list().
func curlDestructuring(arr *syntax.Array) bool {
	var cur syntax.Expr = arr
	for {
		it, ok := cur.Parent().(*syntax.ArrayItem)
		if !ok || it.Value != cur {
			break
		}
		cur = it.Parent().(syntax.Expr) // an Array or a list()
	}
	switch p := cur.Parent().(type) {
	case *syntax.Assign:
		return p.Var == cur
	case *syntax.Foreach:
		return p.Value == cur
	}
	return false
}

// curlClassify implements D4/D5: 1 = enable, -1 = disable, 0 = ignored.
func curlClassify(ctx *analysis.Context, v syntax.Expr, host bool) int {
	want := "1"
	if host {
		want = "2"
	}
	if content, _, ok := astquery.QuotedStringRaw(v); ok {
		return curlBool(content == want)
	}
	if _, ok := v.(*syntax.InterpolatedString); ok {
		return -1
	}
	if c, ok := v.(*syntax.ConstFetch); ok {
		name := astquery.LastNamePart(c.Name.Value)
		switch strings.ToLower(name) {
		case "true":
			return curlBool(!host)
		case "false", "null":
			return -1
		}
		// A constant declared outside this file: classify its value text.
		if k := semanticquery.ResolveConstant(ctx.Types(), c); k != nil {
			return curlClassifyText(strings.TrimSpace(k.Value), want)
		}
		return -1
	}
	return curlClassifyText(ctx.Text(v), want)
}

func curlClassifyText(text, want string) int {
	if len(text) >= 2 && (text[0] == '\'' || text[0] == '"') && text[len(text)-1] == text[0] {
		return curlBool(text[1:len(text)-1] == want)
	}
	if len(text) == 1 {
		return curlBool(text == want)
	}
	switch strings.ToLower(text) {
	case "true":
		return curlBool(want == "1")
	case "false", "null":
		return -1
	}
	return 0
}

func curlBool(enable bool) int {
	if enable {
		return 1
	}
	return -1
}
