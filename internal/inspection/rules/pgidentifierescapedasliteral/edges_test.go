package pgidentifierescapedasliteral

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"$v=pg_escape_literal($c,'x');pg_query(\"SELECT * FROM $v\");", 1},
		{"$v=pg_escape_literal($c,'x');pg_query($c,'SELECT 3');", 0},
		{"$v=pg_escape_literal($c,'x');pg_query($c,\"SELECT $v\");", 0},
		{"$v=pg_escape_literal($c,'x');pg_query($c,\"SELECT * FROM $other\");", 0},
		{"$v=pg_escape_literal($c,'x');pg_query($c,\"SELECT * FROM $v WHERE key=$v\");", 1},
		{"$v=pg_escape_literal($c,'x');pg_query($c,\"SELECT * FROM $v\");echo $v;", 1},
		{"$v=pg_escape_literal($c,'x');$v='y';pg_query($c,\"SELECT * FROM $v\");", 0},
		{"pg_query($c,'SELECT * FROM '.pg_escape_literal($c,'x'));", 1},
		{"$a['x']=pg_escape_literal($c,'x');pg_query($c,\"SELECT * FROM {$a['x']}\");", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PgIdentifierEscapedAsLiteral"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("edges.php", []byte("<?php "+tc.src), syntax.Options{})
			got := e.Analyze(f)
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

func TestIdentifierBoundaries(t *testing.T) {
	for _, tc := range []struct {
		before, after string
		want          bool
	}{
		{"SELECT * FROM ", "", true}, {"SELECT * FROM", "", false}, {"", "", false}, {"SELECT '\"quoted\"' FROM ", "", false}, {"SELECT ", "", false}, {"SELECT * FROM ", ".suffix", false},
	} {
		if got := identifierPosition(tc.before, tc.after); got != tc.want {
			t.Errorf("%q %q: %v", tc.before, tc.after, got)
		}
	}
}

type safetyProbe struct {
	t    *testing.T
	want bool
}

func (safetyProbe) ID() string               { return "PgIdentifierEscapedAsLiteral" }
func (safetyProbe) Semantic()                {}
func (safetyProbe) Flow()                    {}
func (safetyProbe) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (p safetyProbe) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if semanticquery.NativeBuiltin(ctx, c, "pg_escape_literal") {
		if got := safeUses(ctx, c); got != p.want {
			p.t.Errorf("safe uses: got %v want %v", got, p.want)
		}
	}
}

func TestFixSafety(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`$o->value=pg_escape_literal($c,"x");`, false},
		{`$v=pg_escape_literal($c,"x");pg_query($c,"SELECT * FROM $v");$v="later";echo $v;`, true},
		{`$v=pg_escape_literal($c,"x");function unrelated(){echo "later";} pg_query($c,"SELECT * FROM $v");`, true},
		{`$v=pg_escape_literal($c,"x");strlen($v);`, false},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{safetyProbe{t, tc.want}}, analysis.Config{Only: []string{"PgIdentifierEscapedAsLiteral"}})
		if err != nil {
			t.Fatal(err)
		}
		e.Analyze(syntax.Parse("safety.php", []byte("<?php "+tc.src), syntax.Options{}))
	}
}
