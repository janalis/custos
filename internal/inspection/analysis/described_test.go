package analysis_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/meta"
	"github.com/janalis/custos/internal/php/syntax"
)

// ownRule carries its own facts, as a project's custom rule does.
type ownRule struct{ id string }

func (r ownRule) ID() string             { return r.id }
func (ownRule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (ownRule) Fingerprint() string      { return "f1" }
func (r ownRule) Facts() meta.Rule {
	return meta.Rule{ID: r.id, Group: "Custom", Severity: "warning", EnabledByDefault: true}
}
func (r ownRule) Description() meta.Description { return meta.Description{Summary: "own " + r.id} }
func (r ownRule) Check(ctx *analysis.Context, n syntax.Node) {
	if !ctx.KnownRule(r.id) || ctx.KnownRule("Nope") || !ctx.KnownRule("StaticClosureCanBeUsed") {
		panic("known rules")
	}
	ctx.ReportNode(n, "own")
}

// plainRule has neither a catalogue entry nor facts of its own.
type plainRule struct{ ownRule }

func (plainRule) Facts() {}

func TestDescribedRules(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{ownRule{"Own"}}, analysis.Config{})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("a.php", []byte("<?php f(); /** @custos-ignore Own */ g();"), syntax.Options{}))
	if len(got) != 1 || got[0].Rule != "Own" || got[0].Severity != "warning" {
		t.Fatalf("findings: %+v", got)
	}
	if fp := e.Fingerprints(); !slices.Equal(fp, []string{"Own=f1"}) {
		t.Fatalf("fingerprints %v", fp)
	}
	if !e.Known("Own") || !e.Known("StaticClosureCanBeUsed") || e.Known("Other") {
		t.Fatal("Known")
	}
	if _, err := analysis.NewEngine([]analysis.Rule{ownRule{"Own"}, ownRule{"Own"}}, analysis.Config{}); err == nil || !strings.Contains(err.Error(), "registered twice") {
		t.Fatalf("duplicate: %v", err)
	}
	if _, err := analysis.NewEngine([]analysis.Rule{ownRule{"StaticClosureCanBeUsed"}}, analysis.Config{}); err == nil || !strings.Contains(err.Error(), "already a catalogue rule") {
		t.Fatalf("collision: %v", err)
	}
	if _, err := analysis.NewEngine([]analysis.Rule{plainRule{ownRule{"Plain"}}}, analysis.Config{}); err == nil || !strings.Contains(err.Error(), "missing from catalogue") {
		t.Fatalf("unknown: %v", err)
	}
}

func TestLookupDescribed(t *testing.T) {
	rules := []analysis.Rule{plainRule{ownRule{"Plain"}}, ownRule{"Own"}}
	if m, ok := analysis.Lookup(rules, "Own"); !ok || m.Group != "Custom" {
		t.Fatalf("lookup own: %v %v", m, ok)
	}
	if m, ok := analysis.Lookup(rules, "StaticClosureCanBeUsed"); !ok || m.ID != "StaticClosureCanBeUsed" {
		t.Fatal("lookup catalogue")
	}
	if _, ok := analysis.Lookup(rules, "Plain"); ok {
		t.Fatal("lookup plain")
	}
	if d, ok := analysis.Describe(rules, "Own"); !ok || d.Summary != "own Own" {
		t.Fatalf("describe own: %+v", d)
	}
	if d, ok := analysis.Describe(rules, "StaticClosureCanBeUsed"); !ok || d.Summary == "" {
		t.Fatal("describe catalogue")
	}
	if _, ok := analysis.Describe(rules, "Plain"); ok {
		t.Fatal("describe plain")
	}
	if m, ok := analysis.FactsOf(ownRule{"Own"}); !ok || m.ID != "Own" {
		t.Fatal("facts own")
	}
	if m, ok := analysis.FactsOf(ownRule{"StaticClosureCanBeUsed"}); !ok || m.Group == "Custom" {
		t.Fatal("facts catalogue first")
	}
	if _, ok := analysis.FactsOf(plainRule{ownRule{"Plain"}}); ok {
		t.Fatal("facts plain")
	}
}
