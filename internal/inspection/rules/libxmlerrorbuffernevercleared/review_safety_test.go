package libxmlerrorbuffernevercleared

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestReviewCleanupAfterLoop pins which statements may sit between the parse
// loop and the libxml_clear_errors() call that bounds the buffer.
func TestReviewCleanupAfterLoop(t *testing.T) {
	const head = "libxml_use_internal_errors(true);while($more){simplexml_load_string($xml);}"
	for _, tc := range []struct {
		name, tail string
		want       int
	}{
		{"read then clear", "$e=libxml_get_errors();libxml_clear_errors();", 0},
		{"iterate then clear", "foreach(libxml_get_errors() as $e){echo $e->message;}libxml_clear_errors();", 0},
		{"local break and continue", "foreach($a as $e){if($e){continue;}break;}switch($k){case 1:break;}libxml_clear_errors();", 0},
		{"closure body is not searched", "$f=function(){return 1;};$g=fn()=>throw new E();libxml_clear_errors();", 0},
		{"nested declarations are not searched", "function h(){return 1;}class C{function m(){return 1;}}libxml_clear_errors();", 0},
		{"no clear", "$e=libxml_get_errors();", 1},
		{"return first", "if($x){return;}libxml_clear_errors();", 1},
		{"throw first", "$x or throw new E();libxml_clear_errors();", 1},
		{"exit first", "$x or exit(1);libxml_clear_errors();", 1},
		{"goto first", "goto done;libxml_clear_errors();done:", 1},
		{"label first", "again:libxml_clear_errors();", 1},
		{"yield first", "yield 1;libxml_clear_errors();", 1},
		{"yield from first", "yield from $g;libxml_clear_errors();", 1},
		{"counted break in nested loop", "foreach($a as $e){break 2;}libxml_clear_errors();", 1},
		{"counted continue in nested loop", "foreach($a as $e){continue 2;}libxml_clear_errors();", 1},
		{"return in nested loop", "foreach($a as $e){foreach($e as $f){return;}}libxml_clear_errors();", 1},
		{"conditional clear", "if($x){libxml_clear_errors();}", 1},
		{"user clear lookalike", "$o->libxml_clear_errors();", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := reviewCount(t, "function f($more,$xml,$a,$x,$k,$g,$o){"+head+tc.tail+"}"); got != tc.want {
				t.Fatalf("got %d findings want %d", got, tc.want)
			}
		})
	}
	// A break or continue directly after the loop leaves the enclosing iteration.
	for _, escape := range []string{"break", "continue"} {
		src := "while($outer){" + head + "if($x){" + escape + ";}libxml_clear_errors();}"
		if got := reviewCount(t, src); got != 1 {
			t.Fatalf("%s: got %d findings want 1", escape, got)
		}
	}
	if got := reviewCount(t, head+"__halt_compiler();"); got != 1 {
		t.Fatalf("halt: got %d findings want 1", got)
	}
}

func reviewCount(t *testing.T, src string) int {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{New().ID()}})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("review.php", []byte("<?php "+src), syntax.Options{})
	if len(f.Errors) != 0 {
		t.Fatalf("bad source: %+v", f.Errors)
	}
	return len(e.Analyze(f))
}
