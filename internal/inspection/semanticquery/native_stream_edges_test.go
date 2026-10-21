package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

type nativeStreamProbe struct {
	id    string
	check func(*analysis.Context, syntax.Node, string)
}

func (p nativeStreamProbe) ID() string { return p.id }
func (nativeStreamProbe) Semantic()    {}
func (nativeStreamProbe) Flow()        {}
func (nativeStreamProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KForeach, syntax.KReturn}
}
func (p nativeStreamProbe) Check(ctx *analysis.Context, n syntax.Node) { p.check(ctx, n, "proof") }

func TestNativeStreamUnknownContracts(t *testing.T) {
	for _, tc := range []struct {
		id, source string
		check      func(*analysis.Context, syntax.Node, string)
	}{} {
		t.Run(tc.id+"/"+tc.source, func(t *testing.T) {
			probe := nativeStreamProbe{tc.id, tc.check}
			engine, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{tc.id}})
			if err != nil {
				t.Fatal(err)
			}
			file := syntax.Parse("test.php", []byte("<?php "+tc.source), syntax.Options{})
			if len(file.Errors) != 0 {
				t.Fatalf("parse: %+v", file.Errors)
			}
			if findings := engine.Analyze(file); len(findings) != 0 {
				t.Fatalf("unknown facts must not establish a violation: %+v", findings)
			}
		})
	}
}
