package curlheadercallbackmissingbytecount

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

func TestExpansionDCallbackTermination(t *testing.T) {
	for _, tc := range []struct {
		body string
		want bool
	}{
		{"return 'x';", true},
		{"if (true) return 'x';", true},
		{"if (1) { echo 'x'; return 'x'; }", true},
		{"if (false) echo 'x'; else return 'x';", true},
		{"if (false) return 'x';", false},
		{"if ($unknown) return 'x';", false},
		{"if (true) echo 'x'; elseif ($x) return 'x';", false},
		{"echo 'x';", false},
		{strings.Repeat("if (true) {", 65) + "return 'x';" + strings.Repeat("}", 65), false},
	} {
		probeNative(t, "probe(function(){"+tc.body+"});", func(ctx *analysis.Context, c *syntax.FuncCall) {
			body := semanticquery.CallArgument(c.Args, 0, "").(*syntax.Closure).Body
			if got := callbackTerminates(ctx, body); got != tc.want {
				t.Fatalf("%s: got %v want %v", tc.body, got, tc.want)
			}
		})
	}
}

func BenchmarkExpansionDCallbackTermination(b *testing.B) {
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		body := semanticquery.CallArgument(c.Args, 0, "").(*syntax.Closure).Body
		callbackTerminates(ctx, body)
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("bench.php", []byte("<?php probe(function(){if (true) return 'x';});"), syntax.Options{})
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(file)
	}
}

func probeNative(t *testing.T, src string, check func(*analysis.Context, *syntax.FuncCall)) {
	t.Helper()
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		if name, ok := c.Name.(*syntax.Name); ok && name.Value == "probe" {
			check(ctx, c)
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	e.Analyze(syntax.Parse("test.php", []byte("<?php "+src), syntax.Options{}))
}

type nativeProbe struct {
	check func(*analysis.Context, *syntax.FuncCall)
}

func (nativeProbe) ID() string { return "CurlHeaderCallbackMissingByteCount" }

func (nativeProbe) Semantic() {}

func (nativeProbe) Flow() {}

func (nativeProbe) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (p nativeProbe) Check(ctx *analysis.Context, n syntax.Node) { p.check(ctx, n.(*syntax.FuncCall)) }
