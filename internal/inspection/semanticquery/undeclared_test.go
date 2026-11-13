package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

// undeclaredProbe reports every `new` of a class UndeclaredEverywhere
// finds nowhere.
type undeclaredProbe struct{}

func (undeclaredProbe) ID() string               { return "UndefinedClass" }
func (undeclaredProbe) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNew} }
func (undeclaredProbe) Semantic()                {}
func (undeclaredProbe) Check(ctx *analysis.Context, n syntax.Node) {
	name := n.(*syntax.New).Class.(*syntax.Name)
	if UndeclaredEverywhere(ctx, false, ctx.Names().Class(name.Value, name.Span().Start)) {
		ctx.ReportNode(name, "undeclared")
	}
}

func TestUndeclaredEverywhereBlind(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{undeclaredProbe{}}, analysis.Config{Only: []string{"UndefinedClass"}})
	if err != nil {
		t.Fatal(err)
	}
	ix := index.New(stubs.Index())
	ix.SetWorld(index.World{Complete: true, Blind: []string{`app\tests\`}})
	ix.Freeze()
	f := Parse(t, `<?php new \App\Tests\Helper(); new \App\Missing(); new \DateTim(); new \Yac();`)
	if got := e.WithIndex(ix).Analyze(f); len(got) != 2 || Text(f, nodeAt(f, got[0].Span)) != `\App\Missing` {
		t.Fatalf("findings: %+v", got)
	}
}

// nodeAt is a node spanning s, for its text.
func nodeAt(f *syntax.File, s syntax.Span) syntax.Node {
	var found syntax.Node
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if n.Span() == s {
			found = n
		}
		return found == nil
	})
	return found
}
