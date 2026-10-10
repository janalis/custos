// Package varyheaderoverwritesearlierdimensions implements the native VaryHeaderOverwritesEarlierDimensions inspection.
package varyheaderoverwritesearlierdimensions

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Retain the earlier Vary dimensions."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "VaryHeaderOverwritesEarlierDimensions" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	key, value, replace, known := semanticquery.ExpansionHeader(ctx, c)
	if !known || key != "vary" || !replace || value == "*" {
		return
	}
	previous := semanticquery.ExpansionResponseAt(ctx, c).PreviousVary
	if previous == "" {
		return
	}
	current, known := tokens(value)
	if !known {
		return
	}
	old, known := tokens(previous)
	if !known {
		return
	}
	for name := range old {
		if !current[name] {
			ctx.ReportNode(c, message)
			return
		}
	}
}

func tokens(value string) (map[string]bool, bool) {
	result := map[string]bool{}
	for _, part := range strings.Split(value, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			return nil, false
		}
		for _, r := range name {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("!#$%&'*+-.^_`|~", r)) {
				return nil, false
			}
		}
		result[strings.ToLower(name)] = true
	}
	return result, true
}
