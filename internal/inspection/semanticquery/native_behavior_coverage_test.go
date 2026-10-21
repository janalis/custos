package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeSQLLexicalFacts(t *testing.T) {
	for _, tc := range []struct {
		sql                         string
		known                       bool
		markers, positional, quoted int
		identifiers                 bool
	}{
		{"SELECT :id", true, 1, 0, 0, false},
		{"SELECT $$:quoted$$", false, 0, 0, 0, false},
		{"SELECT /* outer /* inner */ */ :id", false, 0, 0, 0, false},
		{"SELECT data ?? 'key' FROM items WHERE id=:id", false, 0, 0, 0, false},
		{"SELECT ? + ?", true, 2, 2, 0, false},
		{"SELECT ':name', :actual", true, 1, 0, 1, false},
		{"SELECT ':missing", false, 0, 0, 0, false},
		{"SELECT /* :ignored ? */ :id -- :ignore ?\n", true, 1, 0, 0, false},
		{"SELECT # :ignored ?\n :id", true, 1, 0, 0, false},
		{"SELECT /* unfinished", false, 0, 0, 0, false},
		{"SELECT 'it''s :literal', `:column`, ':bad name'", true, 0, 0, 0, false},
		{"SELECT 'slash\\quoted', \"double\"", true, 0, 0, 0, false},
		{"SELECT * FROM :table", true, 1, 0, 0, true},
		{"SELECT * FROM things JOIN :joined ON 1=1", true, 1, 0, 0, true},
		{"SELECT CAST(:id AS int)::text", true, 1, 0, 0, false},
		{"UPDATE :table SET label=:label", true, 2, 0, 0, true},
		{"SELECT ':', ':1bad', 'ordinary'", true, 0, 0, 0, false},
	} {
		t.Run(tc.sql, func(t *testing.T) {
			f, known := ScanSQL(tc.sql)
			if known != tc.known || len(f.Markers) != tc.markers || f.Positional != tc.positional || len(f.Quoted) != tc.quoted || f.Identifiers != tc.identifiers {
				t.Fatalf("got %+v known %v, want %+v", f, known, tc)
			}
		})
	}
	for _, tc := range []struct {
		s    string
		want bool
	}{{"", false}, {"name_2", true}, {"2name", false}, {"a-b", false}, {"_name", true}} {
		if got := nativeSQLName(tc.s); got != tc.want {
			t.Fatalf("name %q: got %v, want %v", tc.s, got, tc.want)
		}
	}
}

func TestNativeStatementDominance(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{"earlier(); probe();", true},
		{"if($flag){earlier();} probe();", false},
		{"if(earlier()){probe();}", false},
		{"probe(); earlier();", false},
		{"if($flag){earlier(); probe();}", true},
		{"$flag && earlier(); probe();", false},
		{"$flag || earlier(); probe();", false},
		{"$flag ?? earlier(); probe();", false},
		{"$flag ? earlier() : 0; probe();", false},
		{"earlier() && $flag; probe();", true},
		{"$callback = function(){earlier();}; probe();", false},
		{"$callback = fn()=>earlier(); probe();", false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			var earlier *syntax.FuncCall
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
					if call, ok := n.(*syntax.FuncCall); ok {
						if name, ok := call.Name.(*syntax.Name); ok && name.Value == "earlier" {
							earlier = call
						}
					}
					return true
				})
				if earlier == nil {
					t.Fatal("missing earlier call")
				}
				if got := NativeDominates(earlier, c); got != tc.want {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			})
		})
	}
}

func TestNativeDeterministicCleanupProof(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{"function probe($value){} $h=fopen('path','rb');try{probe($h);}finally{fclose($h);}", true},
		{"$h=fopen('path','rb');try{probe($h);}finally{fclose($h);}", false},
		{"$h=fopen('path','rb');try{probe($h);}finally{echo 'done';}", false},
		{"$h=fopen('path','rb');try{probe($h);}finally{1+2;}", false},
		{"$h=fopen('path','rb');try{probe($h);}finally{fclose($other);}", false},
		{"$h=fopen('path','rb');probe($h);", false},
		{"probe();", false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				if got := CalledInFinally(ctx, c, CallArgument(c.Args, 0, "value"), "fclose"); got != tc.want {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			})
		})
	}
}

func TestNativeCommentPreservation(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{{"probe(1);", true}, {"probe(/*keep*/1);", false}, {"probe(1 //keep\n);", false}, {"probe(1 #keep\n);", false}} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				if got := CommentFreeSpan(ctx, c.Span()); got != tc.want {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			})
		})
	}
}

// nativeCheckProbe exercises the named query API on ordinary parser nodes.
// Checks accept generic nodes and must safely ignore unrelated syntax.
type nativeCheckProbe struct {
	id    string
	check func(*analysis.Context, syntax.Node, string)
}

func (p nativeCheckProbe) ID() string { return p.id }
func (nativeCheckProbe) Semantic()    {}
func (nativeCheckProbe) Flow()        {}
func (nativeCheckProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall, syntax.KEcho, syntax.KReturn, syntax.KForeach, syntax.KBinary, syntax.KAssign, syntax.KIncDec}
}

func (p nativeCheckProbe) Check(ctx *analysis.Context, n syntax.Node) {
	p.check(ctx, n, "Native API contract violation.")
}

func TestNativeProofAuxiliaryBoundaries(t *testing.T) {
	probeNative(t, "echo 'done'; probe('/tmp','/tmp');", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if GlobalStateCall(ctx, c, "committed-output") != nil {
			t.Fatal("echo is not a call state")
		}
		if !SameOperand(ctx, CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")) {
			t.Fatal("equal literal paths should match")
		}
	})
}

func TestNativeNonCallGlobalState(t *testing.T) {
	checked := false
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		if !ctx.IsGlobalFunctionCall(c, "header") {
			return
		}
		checked = true
		if _, known := ctx.Flow().GlobalStateBefore(c, "output"); !known {
			t.Fatal("literal echo must establish output state")
		}
		if GlobalStateCall(ctx, c, "output") != nil {
			t.Fatal("echo state must not resolve to a function call")
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	e.Analyze(syntax.Parse("output.php", []byte("<?php echo 'body'; header('X-Value: example');"), syntax.Options{}))
	if !checked {
		t.Fatal("header was not dispatched")
	}
}
