package semanticquery

import (
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

func TestSQLTokens(t *testing.T) {
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

func TestSQLSkeletonParts(t *testing.T) {
	for _, tc := range []struct {
		src   string
		parts []string
		want  bool
	}{
		{`probe("SELECT '$v'");`, []string{"SELECT '", "'"}, true},
		{`probe('SELECT ' . $v . ' WHERE x=3');`, []string{"SELECT ", " WHERE x=3"}, true},
		{`probe('SELECT 3');`, nil, false},
		{"probe(`echo $v`);", nil, false},
		{`probe($v);`, []string{"", ""}, true},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			p, _, ok := SQLSkeletonParts(ctx, CallArgument(c.Args, 0, ""))
			if ok != tc.want || tc.want && strings.Join(p, "|") != strings.Join(tc.parts, "|") {
				t.Errorf("%s got %v %v", tc.src, p, ok)
			}
		})
	}
}
