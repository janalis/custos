package closurecapturedvaluewrite

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestReviewWorkingLocal pins the working-local exception: a by-value copy
// that the closure itself consumes is not a lost outer update.
func TestReviewWorkingLocal(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"$n=3;$f=function()use($n){while($n-->0){echo '.';}};$f();echo $n;", 0},
		{"$s=' a ';$f=function()use($s){$s=trim($s);echo \"[$s]\";};$f();echo $s;", 0},
		{"$n=0;$f=function()use($n){++$n;echo $n;};$f();echo $n;", 0},
		{"$n=0;$f=function()use($n){echo $n=4;};$f();echo $n;", 0},
		{"$n=0;$f=function()use($n){for($i=0;$i<2;$i++,$n++){}};$f();echo $n;", 0},
		{"$n=0;$f=function()use($n){$n=1;$g=function()use($n){echo $n;};$g();};$f();echo $n;", 0},
		{"$n=0;$f=function()use($n){$n=1;$g=fn()=>$n;};$f();echo $n;", 0},
		{"$n=0;$f=function()use($n){$n=1;log_it($n);};$f();echo $n;", 0},
		// Still lost updates: nothing but the write itself touches the copy.
		{"$n=0;$f=function()use($n){$n=$n+1;};$f();echo $n;", 1},
		{"$s='';$f=function()use($s){$s.='x';};$f();echo $s;", 1},
		{"$n=0;$f=function()use($n){($n++);};$f();echo $n;", 1},
		{"$n=0;$m=0;$f=function()use($m,$n){echo $m;$n--;};$f();echo $n;", 1},
		{"$n=0;$f=function()use($n){$n=1;$g=function(){return 1;};};$f();echo $n;", 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ClosureCapturedValueWrite"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
