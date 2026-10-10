// Package uninitializedtypedpropertyread implements the UninitializedTypedPropertyRead inspection.
package uninitializedtypedpropertyread

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UninitializedTypedPropertyRead" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KPropertyFetch} }

const message = "Initialize this typed property before reading it."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP74 {
		return
	}
	fetch := n.(*syntax.PropertyFetch)
	name, ok := fetch.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	if assign, ok := fetch.Parent().(*syntax.Assign); ok && assign.Var == fetch {
		return
	}
	switch fetch.Parent().(type) {
	case *syntax.Isset, *syntax.Empty:
		return
	}
	if binary, ok := fetch.Parent().(*syntax.Binary); ok && binary.Left == fetch && binary.Op.Kind == syntax.TCoalesce {
		return
	}
	receiver, ok := semanticquery.NativeValue(ctx, fetch.Var).(*syntax.New)
	if !ok {
		return
	}
	cls := ctx.Types().ClassRef(receiver.Class)
	if cls == "" || !ctx.Index().AncestorsComplete(cls, ctx.PHP) || !semanticquery.HierarchyResolved(ctx.Index(), cls, ctx.PHP) || ctx.Index().FindMethod(cls, "__construct", ctx.PHP) != nil || ctx.Index().FindMethod(cls, "__get", ctx.PHP) != nil {
		return
	}
	property := ctx.Index().FindProperty(cls, name.Value, ctx.PHP)
	if property == nil || property.Type == "" || property.HasDefault || property.Magic || property.ReadsRunCode {
		return
	}
	// Restrict proof to fresh object expressions: aliases may have been initialized.
	if _, ok := syntax.UnwrapParens(fetch.Var).(*syntax.New); !ok {
		return
	}
	ctx.ReportNode(fetch, message)
}

func (rule) Semantic() {}
