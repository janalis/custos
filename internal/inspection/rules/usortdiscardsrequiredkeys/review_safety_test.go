package usortdiscardsrequiredkeys

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, Only: []string{"UsortDiscardsRequiredKeys"}})
	if err != nil {
		t.Fatal(err)
	}
	const setup = "<?php $a=['k'=>2,'j'=>1];"
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"coalesce assignment", "usort($a,$cmp);$a['k']??=5;", 0},
		{"isset guarded body", "usort($a,$cmp);if(isset($a['k'])){echo $a['k'];}", 0},
		{"isset guarded ternary", "usort($a,$cmp);echo isset($a['k'])?$a['k']:0;", 0},
		{"empty guarded operand", "usort($a,$cmp);echo !empty($a['k'])&&$a['k']>1;", 0},
		{"coalesce guarded body", "usort($a,$cmp);if(($a['k']??null)!==null){echo $a['k'];}", 0},
		{"array_key_exists guard", "usort($a,$cmp);if(array_key_exists('k',$a)){echo $a['k'];}", 0},
		{"key_exists named guard", "usort($a,$cmp);if(key_exists(array:$a,key:'k')){echo $a['k'];}", 0},
		{"write then read", "usort($a,$cmp);if($f){$a['k']=1;echo $a['k'];}", 0},
		{"unset then read", "usort($a,$cmp);if($f){unset($a['k']);echo $a['k'];}", 0},
		{"replace then read", "usort($a,$cmp);if($f){$a=['k'=>1];echo $a['k'];}", 0},
		{"append only", "usort($a,$cmp);$a[]=3;", 0},
		{"guard on other key", "usort($a,$cmp);if(isset($a['j'])){echo $a['k'];}", 1},
		{"guard on other array", "usort($a,$cmp);if(array_key_exists('k',$b)){echo $a['k'];}", 1},
		{"guard with unknown key", "usort($a,$cmp);if(array_key_exists($x,$a)){echo $a['k'];}", 1},
		{"unrelated call", "usort($a,$cmp);if(count($a)){echo $a['k'];}", 1},
		{"read then write", "usort($a,$cmp);if($f){echo $a['k'];$a['k']=1;}", 1},
		{"read then replace", "usort($a,$cmp);if($f){echo $a['k'];$a=[];}", 1},
		{"read feeding its own write", "usort($a,$cmp);$a['k']=$a['k']+1;", 1},
		{"compound update", "usort($a,$cmp);$a['k'].='x';", 1},
		{"read in assigned value", "usort($a,$cmp);$b['x']=$a['k'];", 1},
		{"read as dynamic key", "usort($a,$cmp);echo $a[$a['k']];", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(setup + tc.source)
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
			res := fixing.FixSource(e, "review.php", src, fixing.Options{Parse: opt, SinglePass: true})
			want := setup + tc.source
			if tc.findings != 0 {
				want = setup + "\\uasort" + tc.source[len("usort"):]
			}
			if string(res.Source) != want {
				t.Fatalf("fixed source: got %s, want %s", res.Source, want)
			}
		})
	}
}
