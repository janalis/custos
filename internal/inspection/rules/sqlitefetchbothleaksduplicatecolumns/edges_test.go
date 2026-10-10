package sqlitefetchbothleaksduplicatecolumns

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
		{`function f(SQLite3Result $r){$row=$r->fetchArray();echo $row[0];json_encode($row);}`, 0},
		{`function f(SQLite3Result $r){$row=$r->fetchArray();json_encode($row);echo $row[0];}`, 0},
		{`function f(SQLite3Result $r){$row=$r->fetchArray();function unrelated(){}json_encode($row);}`, 1},
		{`function f(SQLite3Result $r){$a['row']=$r->fetchArray();json_encode($a['row']);}`, 0},
		{`function f(SQLite3Result $r){$row=$r->fetchArray();json_encode($row);$row=[];echo $row;}`, 1},

		{"function f(SQLite3Result $r){strlen('x');json_encode([]);json_encode($r->fetchArray());}", 1},
		{"function f(SQLite3Result $r){json_encode($r->fetchArray($mode));}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SqliteFetchBothLeaksDuplicateColumns"}})
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

type safetyProbe struct {
	t    *testing.T
	want bool
}

func (safetyProbe) ID() string               { return "SqliteFetchBothLeaksDuplicateColumns" }
func (safetyProbe) Semantic()                {}
func (safetyProbe) Flow()                    {}
func (safetyProbe) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (p safetyProbe) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if semanticquery.NativeMethod(ctx, c, "SQLite3Result", "fetchArray") {
		if got := jsonOnly(ctx, c); got != p.want {
			p.t.Errorf("jsonOnly got %v want %v", got, p.want)
		}
	}
}

func TestFixConsumers(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`function f(SQLite3Result $r){$a['row']=$r->fetchArray();}`, false},
		{`$d=new SQLite3(':memory:');$r=$d->query('SELECT 1');$row=$r->fetchArray();json_encode($row);`, true},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{safetyProbe{t, tc.want}}, analysis.Config{Only: []string{"SqliteFetchBothLeaksDuplicateColumns"}})
		if err != nil {
			t.Fatal(err)
		}
		e.Analyze(syntax.Parse("safe.php", []byte("<?php "+tc.src), syntax.Options{}))
	}
}
