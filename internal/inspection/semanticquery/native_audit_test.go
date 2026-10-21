package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func auditNative(t *testing.T, id string, check func(*analysis.Context, syntax.Node, string), src string) []diagnostic.Finding {
	t.Helper()
	p := nativeCheckProbe{id, check}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{id}})
	if err != nil {
		t.Fatal(err)
	}
	return e.Analyze(syntax.Parse("audit.php", []byte("<?php "+src), syntax.Options{}))
}

func TestNativeAuditIntentAndIdentity(t *testing.T) {
	for _, tc := range []struct {
		id, src string
		check   func(*analysis.Context, syntax.Node, string)
		want    int
	}{} {
		t.Run(tc.id+tc.src, func(t *testing.T) {
			got := auditNative(t, tc.id, tc.check, tc.src)
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}

func TestNativeAuditFixSyntax(t *testing.T) {
	for _, tc := range []struct {
		id, src, needle string
		check           func(*analysis.Context, syntax.Node, string)
		fixes           int
	}{} {
		t.Run(tc.id+tc.src, func(t *testing.T) {
			got := auditNative(t, tc.id, tc.check, tc.src)
			if len(got) != 1 || len(got[0].Fixes) != tc.fixes {
				t.Fatalf("got %+v, want one with %d fixes", got, tc.fixes)
			}
			if tc.fixes > 0 {
				edits := got[0].Fixes[0].Edits()
				found := false
				for _, edit := range edits {
					if strings.Contains(edit.NewText, tc.needle) {
						found = true
					}
				}
				if !found {
					t.Fatalf("edits %+v lack %q", edits, tc.needle)
				}
				if len(edits) != 1 {
					t.Fatalf("expected one reviewable edit, got %+v", edits)
				}
				edit := edits[0]
				original := "<?php " + tc.src
				fixed := original[:edit.Span.Start] + edit.NewText + original[edit.Span.End:]
				if parsed := syntax.Parse("fixed.php", []byte(fixed), syntax.Options{}); len(parsed.Errors) > 0 {
					t.Fatalf("fix does not parse: %s: %+v", fixed, parsed.Errors)
				}
			}
		})
	}
}

func TestNativeAuditSharedProofBoundaries(t *testing.T) {
	probeNative(t, "$x='same';probe($x,$x);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if !SameOperand(ctx, CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")) {
			t.Fatal("uniquely reaching assigned value lost")
		}
		if SameOperand(ctx, nil, CallArgument(c.Args, 0, "")) {
			t.Fatal("missing path must be unknown")
		}
	})
	probeNative(t, "probe($unknown,$unknown);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if SameOperand(ctx, CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")) {
			t.Fatal("unknown variables cannot prove identity")
		}
	})
	probeNative(t, "probe(mb_strlen($body,chooseEncoding()));", func(_ *analysis.Context, c *syntax.FuncCall) {
		call := CallArgument(c.Args, 0, "").(*syntax.FuncCall)
		if ArgumentsRemovable(call) {
			t.Fatal("effectful optional argument must remain evaluated")
		}
	})
}
