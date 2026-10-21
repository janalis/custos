// Package curlwholeurlescapedascomponent implements CurlWholeUrlEscapedAsComponent.
package curlwholeurlescapedascomponent

import (
	"net/url"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Escape URL components before assembling the URL."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlWholeUrlEscapedAsComponent" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	value := semanticquery.CurlSetting(ctx, c, "CURLOPT_URL")
	origin := semanticquery.OriginCall(ctx, value, "curl_escape")
	if origin == nil {
		return
	}
	text, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(origin.Args, 1, "string"))
	if known && expansionDAbsoluteURL(text) {
		ctx.ReportNode(c, message)
	}
}

// Keep URL parsing bound to Go's parser; PHP contracts are checked separately.
func expansionDAbsoluteURL(text string) bool {
	u, err := url.Parse(text)
	return err == nil && u.IsAbs() && u.Host != "" && strings.Contains(text, "://")
}
