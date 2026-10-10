package loopclosurecapturesreference

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "LoopClosureCapturesReference" || len(r.Kinds()) == 0 {
		t.Fatal("invalid rule contract")
	}
	if s, ok := r.(interface{ Semantic() }); ok {
		s.Semantic()
	}
	if f, ok := r.(interface{ Flow() }); ok {
		f.Flow()
	}
}

func TestProofBoundaries(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{{"$f=function(){};", 0}, {"$a[0]=function(){};", 0}, {"foo(function(){});", 0}, {"$jobs=[]; for($i=0;$i<3;$i++) $jobs[]=function()use(&$i){return $i;};", 0}, {"$jobs=[]; for(;$i<3;$i++){ $jobs[]=function()use(&$i){return $i;}; }", 0}, {"$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function()use(&$i){return $i;}; echo 1; }", 0}, {"$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function()use(&$i){return $i;}; }", 0}, {"$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function()use(&$i){return $i;}; } echo 1;", 0}, {"$jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function()use(&$i){return $i;}; } foreach($jobs as $job){$f=function(){return 1;};echo 1;}", 0}, {"for($i=0;$i<3;$i++){foo($jobs[]=function()use(&$i){return $i;});}", 0}, {"for(++$i;$i<3;$i++){$jobs[]=function()use(&$i){return $i;};}", 0}, {"for($obj->p=0;$i<3;$i++){$jobs[]=function()use(&$i){return $i;};}", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"LoopClosureCapturesReference"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
