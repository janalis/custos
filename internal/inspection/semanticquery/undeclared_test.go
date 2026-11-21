package semanticquery

import (
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/index"
	"github.com/janalis/custos/internal/semantic/stubs"
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

// TestUndeclaredEverywhereHosted: a file importing from a namespace whose
// package loads a bundle of other code (PHPStan's phar) runs within that
// package, which may provide any name.
func TestUndeclaredEverywhereHosted(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{undeclaredProbe{}}, analysis.Config{Only: []string{"UndefinedClass"}})
	if err != nil {
		t.Fatal(err)
	}
	ix := index.New(stubs.Index())
	ix.SetWorld(index.World{Complete: true, Hosts: []string{`phpstan\`}})
	ix.Freeze()
	for src, want := range map[string]int{
		`<?php namespace App; use PHPStan\Type\Type; new \PhpParser\Node\Identifier();`:                                                     0,
		`<?php namespace App; use PHPStan\{Type\Type, Analyser\Scope}; new \PhpParser\Node\Name();`:                                         0,
		`<?php namespace App; use Other\Type; new \PhpParser\Node\Name();`:                                                                  1,
		`<?php namespace PHPStan\Ext; new \PhpParser\Node\Name();`:                                                                          0,
		`<?php namespace App { final class R implements \PHPStan\Rules\Rule { function f() { return new \PhpParser\Node\Name(); } } }`:      0,
		`<?php namespace App; abstract class R extends Type implements \Other\Rule { function f() { return new \PhpParser\Node\Name(); } }`: 1,
		`<?php namespace App; use PHPStanned\Type; new \PhpParser\Node\Name();`:                                                             1,
	} {
		f := Parse(t, src)
		if got := e.WithIndex(ix).Analyze(f); len(got) != want {
			t.Errorf("%s: %d findings, want %d", src, len(got), want)
		}
	}
}
