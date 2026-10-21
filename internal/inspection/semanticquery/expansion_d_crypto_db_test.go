package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestExpansionCryptoDBSQLGrammar(t *testing.T) {
	for _, tc := range []struct {
		sql   string
		known bool
	}{
		{"SELECT 1 -- :ignored\n, :id # :ignored\n /* :ignored */", true},
		{"SELECT 'a\\b', 'a''b', `label`, 1::int", true},
		{"SELECT $body$", false},
		{"SELECT /* unfinished", false},
		{"SELECT /* /* nested */", false},
		{"SELECT 'unfinished", false},
		{"SELECT 'unfinished\\", false},
		{strings.Repeat("a", 65537), false},
	} {
		_, known := SQLTokens(tc.sql)
		if known != tc.known {
			t.Fatalf("%q: known=%v", tc.sql, known)
		}
	}
	tokens, known := SQLTokens("SELECT ':x', :x, :X /* :x */")
	if !known || strings.Join(tokens, " ") != "SELECT ':x , :x , :X" {
		t.Fatalf("incorrect marker tokens: %v", tokens)
	}
}

func TestExpansionCryptoDBLocalProof(t *testing.T) {
	for _, tc := range []struct {
		src  string
		same bool
	}{
		{"probe(1,1);", true},
		{"probe(1,2);", false},
		{"probe($a,$b);", false},
		{"function f($a){useValue($a);unset($a);probe($a);}", false},
		{"function f($a){useValue($a);$o->unknown($a);probe($a);}", false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			first := CallArgument(c.Args, 0, "a")
			second := CallArgument(c.Args, 1, "b")
			if second == nil {
				for _, ev := range ScopeCallEvents(ctx, c) {
					if prior, ok := ev.(*syntax.FuncCall); ok && ctx.Text(prior.Name) == "useValue" {
						second = CallArgument(prior.Args, 0, "")
					}
				}
			}
			if same := SameStableLocal(ctx, first, second); same != tc.same {
				t.Fatalf("%s: same=%v", tc.src, same)
			}
			if SameStableLocal(ctx, first, nil) {
				t.Fatal("nil proves identity")
			}
		})
	}
	probeNative(t, "probe(1);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		for _, x := range []syntax.Expr{nil, CallArgument(c.Args, 0, "")} {
			if _, ok := KnownByteLength(ctx, x); ok {
				t.Fatal("non-string has known length")
			}
		}
	})
}

func TestExpansionCryptoDBBudgetsAndSuccessGuards(t *testing.T) {
	probeNative(t, "function f($a){"+strings.Repeat("other();", 513)+"probe($a,$a);}", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if SameStableLocal(ctx, CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")) {
			t.Fatal("exhausted scope establishes identity")
		}
	})
}
