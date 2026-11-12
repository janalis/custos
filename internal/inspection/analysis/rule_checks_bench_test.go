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
