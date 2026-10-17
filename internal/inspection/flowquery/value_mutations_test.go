package flowquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

func TestValueDiscoveryRejectsIndirectWrites(t *testing.T) {
	for _, body := range []string{
		`$x=17; mutate($x); probe($x);`,
		`$x=17; $object->mutate($x); probe($x);`,
		`$x=17; new Writer($x); probe($x);`,
		`$x=17; $alias=&$x; mutate($alias); probe($x);`,
		`$x=17; $aliases=[&$x]; mutate($aliases); probe($x);`,
		`$x=[17]; foreach($x as &$alias){$alias='changed';} probe($x);`,
		`$x=17; while($again){probe($x); mutate($x);}`,
		`$x=17; $capture=function()use(&$x){}; invoke($capture); probe($x);`,
		`$x=17; mutate($$name); probe($x);`,
		`$x=[17]; mutate($x[0]); probe($x);`,
	} {
		t.Run(body, func(t *testing.T) {
			f := Parse(t, "<?php function example($object,$again,$name){"+body+"}")
			var read syntax.Expr
			syntax.InspectFile(f, func(n syntax.Node) bool {
				if call, ok := n.(*syntax.FuncCall); ok && astquery.CallLastName(call) == "probe" {
					read = call.Args.Args[0].(*syntax.Arg).Value
				}
				return true
			})
			for _, discover := range []func(*syntax.File, syntax.Expr) ([]syntax.Expr, bool){PossibleValuesKnown, PossibleValuesComplete, PossibleValuesReaching} {
				if values, known := discover(f, read); known || len(values) != 0 {
					t.Fatalf("indirectly written value claimed known: %v %v", values, known)
				}
			}
		})
	}
}

func BenchmarkValueDiscoveryIndirectWrites(b *testing.B) {
	src := `<?php function example(){ $x=17;` + strings.Repeat(`read($x);`, 100) + `probe($x); }`
	f := syntax.Parse("values.php", []byte(src), syntax.Options{})
	var read syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if call, ok := n.(*syntax.FuncCall); ok && astquery.CallLastName(call) == "probe" {
			read = call.Args.Args[0].(*syntax.Arg).Value
		}
		return true
	})
	byValue := func(*syntax.Arg) bool { return false }
	b.ReportAllocs()
	for b.Loop() {
		CompleteValuesWithReferences(f, read, byValue)
	}
}

func TestValueDiscoveryReadBeforeArgumentWrite(t *testing.T) {
	if got := completeValuesOf(t, `<?php function example(){ $x=17; probe($x); mutate($x); }`); got != "17" {
		t.Fatal(got)
	}
	f := Parse(t, `<?php function example(){ $x=17; probe($x); }`)
	call := f.Stmts[0].(*syntax.Function).Body.Stmts[1].(*syntax.ExprStmt).Expr.(*syntax.FuncCall)
	if len(PossibleValues(f, call.Args.Args[0].(*syntax.Arg).Value)) != 1 {
		t.Fatal("plain possible-values wrapper lost the assignment")
	}
}

func TestIndirectWriteWorkBound(t *testing.T) {
	f := Parse(t, `<?php function example(){ $x=17; consume($x); probe($x); }`)
	fn := f.Stmts[0].(*syntax.Function)
	prior := fn.Body.Stmts[1].(*syntax.ExprStmt).Expr.(*syntax.FuncCall).Args.Args[0].(*syntax.Arg)
	read := fn.Body.Stmts[2].(*syntax.ExprStmt).Expr.(*syntax.FuncCall).Args.Args[0].(*syntax.Arg).Value.(*syntax.Variable)
	w := NewValueWalk(f)
	w.ReferenceArgument = func(*syntax.Arg) bool { return false }
	ix := &assignIndex{arguments: map[string][]*syntax.Arg{"x": {}}}
	for range MaxAssignScan + 1 {
		ix.arguments["x"] = append(ix.arguments["x"], prior)
	}
	if !w.indirectWrite(ix, read) {
		t.Fatal("exhausted argument budget established a value proof")
	}
	ix.arguments["x"] = []*syntax.Arg{{}}
	if !w.indirectWrite(ix, read) {
		t.Fatal("an unlinked argument established a value proof")
	}
	f = Parse(t, `<?php function example(){ $x=17; probe($x); if($flag){consume($x);} }`)
	fn = f.Stmts[0].(*syntax.Function)
	read = fn.Body.Stmts[1].(*syntax.ExprStmt).Expr.(*syntax.FuncCall).Args.Args[0].(*syntax.Arg).Value.(*syntax.Variable)
	var later *syntax.Arg
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if call, ok := n.(*syntax.FuncCall); ok && astquery.CallLastName(call) == "consume" {
			later = call.Args.Args[0].(*syntax.Arg)
		}
		return true
	})
	ix.arguments["x"] = nil
	for range MaxAssignScan {
		ix.arguments["x"] = append(ix.arguments["x"], later)
	}
	if !w.indirectWrite(ix, read) {
		t.Fatal("exhausted ancestor budget established a value proof")
	}
}
