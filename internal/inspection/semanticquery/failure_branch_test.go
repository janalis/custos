package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func failureProbe(t testing.TB, condition, body string) *syntax.FuncCall {
	t.Helper()
	f := syntax.Parse("proof.php", []byte("<?php function f(){if("+condition+"){"+body+"}}"), syntax.Options{})
	if len(f.Errors) > 0 {
		t.Fatal(f.Errors)
	}
	var call *syntax.FuncCall
	syntax.Inspect(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if name, ok := c.Name.(*syntax.Name); ok && name.Value == "probe" {
				call = c
			}
		}
		return true
	})
	if call == nil {
		t.Fatal("missing probe")
	}
	return call
}

func TestFailureBranchProof(t *testing.T) {
	for _, tc := range []struct {
		name, condition, body string
		want                  bool
	}{
		{"negated", "!probe()", "return false;", true},
		{"plain", "probe()", "return false;", false},
		{"even", "!!probe()", "throw new Exception();", false},
		{"odd", "!!!probe()", "throw new Exception();", true},
		{"mixed even", "!(probe()==false)", "return false;", false},
		{"reverse equality", "false==probe()", "return false;", true},
		{"numeric equality", "probe()==0", "return false;", false},
		{"strict", "probe()===false", "return false;", false},
		{"suppressed assignment", "!(@($r=probe()))", "return false;", true},
		{"compound", "!($r+=probe())", "return false;", false},
		{"reference", "!($r=&probe())", "return false;", false},
		{"assignment target", "!($a[probe()]=false)", "return false;", false},
		{"other unary", "+probe()", "return false;", false},
		{"conditional", "!probe()", "if($strict){throw new Exception();}", false},
		{"strict guard", "!($r=probe())", "if($r===false){throw new Exception();}", false},
		{"both reject", "!probe()", "if($x){throw new Exception();}else{return false;}", true},
		{"elseif reject", "!probe()", "if($x){return false;}elseif($y){return false;}else{throw new Exception();}", true},
		{"elseif falls", "!probe()", "if($x){return false;}elseif($y){return false;}", false},
		{"later reject", "!probe()", "if($x){return false;}logEvent();;return false;", true},
		{"other return", "!probe()", "if($x){return true;}throw new Exception();", false},
		{"callback", "!probe()", "$f=function(){yield 1;return true;};return false;", true},
		{"callback only", "!probe()", "$f=function(){throw new Exception();};", false},
		{"loop", "!probe()", "while($x){return false;}", false},
		{"try", "!probe()", "try{throw new Exception();}catch(Exception $e){}", false},
		{"yield", "!probe()", "yield 1;return false;", false},
		{"yield from", "!probe()", "yield from $items;return false;", false},
		{"jump", "!probe()", "goto done;return false;done:;", false},
		{"label", "!probe()", "return false;done:;", false},
		{"exit", "!probe()", "exit();return false;", false},
		{"embedded throw", "!probe()", "$x=$y??throw new Exception();return false;", false},
		{"unreachable statement", "!probe()", "return false;work();", true},
		{"condition throws", "!probe()", "if($x??throw new Exception()){return false;}else{return false;}", false},
		{"elseif throws", "!probe()", "if($x){return false;}elseif($y??throw new Exception()){return false;}else{return false;}", false},
		{"outcome budget", "!probe()", strings.Repeat(";", 150) + "return false;", false},
		{"repeated expression budget", "!probe()", "$x=[" + strings.Repeat("1,", 100) + "];return false;", false},
		{"body budget", "!probe()", strings.Repeat("work();", 300) + "return false;", false},
		{"expression budget", "!probe()", "$x=[" + strings.Repeat("1,", 300) + "];return false;", false},
		{"condition budget", strings.Repeat("!", 257) + "probe()", "return false;", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			call := failureProbe(t, tc.condition, tc.body)
			if got := FailureBranch(call) != nil; got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func BenchmarkFailureBranchProof(b *testing.B) {
	call := failureProbe(b, "!probe()", "if($x){return false;}else{throw new Exception();}")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		FailureBranch(call)
	}
}
