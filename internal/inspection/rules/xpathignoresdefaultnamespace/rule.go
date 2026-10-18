// Package xpathignoresdefaultnamespace implements the native XPathIgnoresDefaultNamespace inspection.
package xpathignoresdefaultnamespace

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Register a namespace prefix for the XPath query."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "XPathIgnoresDefaultNamespace" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "DOMXPath", "query") {
		return
	}
	query, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "expression"))
	if !known || !strings.HasPrefix(query, "//") {
		return
	}
	target := strings.TrimPrefix(query, "//")
	if target == "" || strings.ContainsAny(target, "/:[]*@() .-") {
		return
	}
	creation, ok := semanticquery.NativeValue(ctx, c.Var).(*syntax.New)
	if !ok || !semanticquery.NamesGlobalClass(ctx, creation.Class, "DOMXPath") {
		return
	}
	// The three evidence statements must be adjacent: anything between the
	// load and the query could mutate or replace the document.
	prior := flowquery.NativePriorStatements(ctx.File, c)
	if len(prior) < 2 {
		return
	}
	loading, ok := prior[len(prior)-2].(*syntax.ExprStmt)
	if !ok {
		return
	}
	st, ok := prior[len(prior)-1].(*syntax.ExprStmt)
	if !ok {
		return
	}
	assignment, ok := st.Expr.(*syntax.Assign)
	if !ok || assignment.Value != creation {
		return
	}
	document := semanticquery.CallArgument(creation.Args, 0, "document")
	calls, historyKnown := semanticquery.NativeCallHistory(ctx, creation, document, "loadxml").Certain()
	if !historyKnown {
		return
	}
	if len(calls) == 0 {
		return
	}
	if load, ok := calls[len(calls)-1].(*syntax.MethodCall); ok && syntax.Node(load) == syntax.Node(loading.Expr) && semanticquery.NativeMethod(ctx, load, "DOMDocument", "loadXML") {
		text, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(load.Args, 0, "source"))
		if !known || len(text) > 65536 {
			return
		}
		if missesNamespace(text, target) {
			ctx.ReportNode(c, message)
		}
	}
}

func missesNamespace(text, target string) bool {
	decoder := xml.NewDecoder(strings.NewReader(text))
	found := false
	root := true
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return found
		}
		if err != nil {
			return false
		}
		if element, ok := token.(xml.StartElement); ok {
			if root {
				root = false
				if element.Name.Space == "" {
					return false
				}
			}
			if element.Name.Local == target {
				if element.Name.Space == "" {
					return false
				}
				found = true
			}
		}
	}
}
