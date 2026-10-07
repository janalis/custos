package analysis

import (
	"testing"

	"custos/internal/phpver"
	"custos/internal/syntax"
)

type crashRule struct{}

func (crashRule) ID() string               { return "UnnecessarySemicolon" }
func (crashRule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNop} }
func (crashRule) Check(*Context, syntax.Node) {
	var p *int
	_ = *p
}

type okRule struct{}

func (okRule) ID() string                        { return "NestedNotOperators" }
func (okRule) Kinds() []syntax.NodeKind          { return []syntax.NodeKind{syntax.KUnary} }
func (okRule) Check(ctx *Context, n syntax.Node) { ctx.ReportNode(n, "found") }

func TestRulePanicIsolated(t *testing.T) {
	e, err := NewEngine([]Rule{crashRule{}, okRule{}}, Config{EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("x.php", []byte("<?php ;\n$a = !$b;"), syntax.Options{Version: phpver.PHP84})
	got := e.Analyze(f)
	if len(got) != 2 || got[0].Rule != "internal" || got[1].Rule != "NestedNotOperators" {
		t.Fatalf("got %+v", got)
	}
}
