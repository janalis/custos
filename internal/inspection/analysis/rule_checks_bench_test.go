package analysis

import (
	"testing"

	"custos/internal/inspection/meta"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// idRule is a rule of the catalogue that checks nothing.
type idRule struct{ id string }

func (r idRule) ID() string                { return r.id }
func (idRule) Kinds() []syntax.NodeKind    { return nil }
func (idRule) Check(*Context, syntax.Node) {}

// BenchmarkRuleChecks asks, with every catalogue rule enabled, whether
// the last one checks a node, as a rule deferring to it does per check.
func BenchmarkRuleChecks(b *testing.B) {
	all, err := meta.All()
	if err != nil {
		b.Fatal(err)
	}
	rules := make([]Rule, 0, len(all))
	for _, m := range all {
		rules = append(rules, idRule{m.ID})
	}
	e, err := NewEngine(rules, Config{EnableAll: true})
	if err != nil {
		b.Fatal(err)
	}
	f := syntax.Parse("x.php", []byte("<?php $a = !$b;"), syntax.Options{Version: phpversion.PHP84})
	ctx := &Context{File: f, Src: f.Src, PHP: e.cfg.PHP, engine: e}
	last := all[len(all)-1].ID
	b.ReportAllocs()
	for b.Loop() {
		if !ctx.RuleChecks(last, f) {
			b.Fatal("not enabled")
		}
	}
}

// BenchmarkRuleChecksPerNode asks, in one run over a file, whether a rule
// checks each of its nodes, as a rule deferring per node does.
func BenchmarkRuleChecksPerNode(b *testing.B) {
	e, err := NewEngine([]Rule{idRule{"ImpossibleTypeComparison"}}, Config{EnableAll: true})
	if err != nil {
		b.Fatal(err)
	}
	src := []byte("<?php\n")
	for range 200 {
		src = append(src, "$a = $b === $c && f($d, $e[1]);\n"...)
	}
	f := syntax.Parse("x.php", src, syntax.Options{Version: phpversion.PHP84})
	var nodes []syntax.Node
	syntax.InspectFile(f, func(n syntax.Node) bool { nodes = append(nodes, n); return true })
	b.ReportAllocs()
	for b.Loop() {
		ctx := &Context{File: f, Src: f.Src, PHP: e.cfg.PHP, engine: e}
		for _, n := range nodes {
			if !ctx.RuleChecks("ImpossibleTypeComparison", n) {
				b.Fatal("not enabled")
			}
		}
	}
}
