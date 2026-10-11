// Package uninitializedtypedpropertyread implements the UninitializedTypedPropertyRead inspection.
package uninitializedtypedpropertyread

import (
	"strings"

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
	// A guard covers the whole access chain rooted at the property: an element
	// or sub-property read inside isset/empty/?? never throws. Calls end the chain.
	var guarded syntax.Node = fetch
	for chained(guarded) {
		guarded = guarded.Parent()
	}
	switch parent := guarded.Parent().(type) {
	case *syntax.Isset, *syntax.Empty:
		return
	case *syntax.Binary:
		if parent.Left == guarded && parent.Op.Kind == syntax.TCoalesce {
			return
		}
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
	// `new static` may build a subclass whose constructor initializes the property.
	if ref, ok := receiver.Class.(*syntax.Name); ok && strings.EqualFold(ref.Value, "static") {
		if decl := ctx.Index().Class(cls, ctx.PHP); decl == nil || !decl.Final {
			return
		}
	}
	// Restrict proof to fresh object expressions: aliases may have been initialized.
	if _, ok := syntax.UnwrapParens(fetch.Var).(*syntax.New); !ok {
		return
	}
	ctx.ReportNode(fetch, message)
}

// chained reports whether n is the base of an enclosing element or property
// access (or a parenthesized form of one), so a guard around the parent also
// guards n. Method calls and index positions do not qualify.
func chained(n syntax.Node) bool {
	switch parent := n.Parent().(type) {
	case *syntax.ArrayDimFetch:
		return parent.Var == n
	case *syntax.PropertyFetch:
		return parent.Var == n
	case *syntax.Paren:
		return true
	}
	return false
}

func (rule) Semantic() {}
