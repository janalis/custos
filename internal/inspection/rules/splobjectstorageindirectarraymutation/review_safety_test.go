package splobjectstorageindirectarraymutation

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestReviewReplacedValue covers writes that may replace the stored array
// between the proving assignment and the nested mutation.
func TestReviewReplacedValue(t *testing.T) {
	const head = "$s=new SplObjectStorage();$o=new stdClass();$s[$o]=['c'=>1];"
	for _, tc := range []struct {
		name string
		src  string
		want int
	}{
		{"offsetSet", "$s->offsetSet($o,new ArrayObject());$s[$o]['c']=2;", 0},
		{"attach", "$s->attach($o,new ArrayObject());$s[$o]['c']=2;", 0},
		{"addAll", "$s->addAll($other);$s[$o]['c']=2;", 0},
		{"dynamic method", "$s->$m($o);$s[$o]['c']=2;", 0},
		{"aliased receiver", "$t=$s;$t->attach($o,new ArrayObject());$s[$o]['c']=2;", 0},
		{"conditional object", "if($c){$s[$o]=new ArrayObject();}$s[$o]['c']=2;", 0},
		{"other key object", "$s[$p]=new ArrayObject();$s[$o]['c']=2;", 0},
		{"direct unset", "unset($s[$o]);$s[$o]['c']=2;", 0},
		{"capturing closure", "$f=function() use($s,$o){$s[$o]=new ArrayObject();};$f();$s[$o]['c']=2;", 0},
		{"capturing arrow function", "$f=fn()=>$s[$o]=new ArrayObject();$f();$s[$o]['c']=2;", 0},
		{"unrelated arrow function", "$f=fn($x)=>$x+1;$s[$o]['c']=2;", 1},
		{"top-level later write", "$s[$o]['c']=2;$s->attach($o,new ArrayObject());", 1},
		{"read-only call", "$s->contains($o);$s[$o]['c']=2;", 1},
		{"array restored", "$s->detach($o);$s[$o]=[];$s[$o]['c']=2;", 1},
		{"conditional array", "if($c){$s[$o]=['d'=>1];}$s[$o]['c']=2;", 1},
		{"other key array", "$s[$p]=[];$s[$o]['c']=2;", 1},
		{"unrelated closure", "$f=function() use($o){};$s[$o]['c']=2;", 1},
		{"unrelated call", "$q->attach($o);$s[$o]['c']=2;", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SplObjectStorageIndirectArrayMutation"}})
			if err != nil {
				t.Fatal(err)
			}
			for _, src := range []string{"<?php " + head + tc.src, "<?php function f($c,$p,$q,$m,$other){" + head + tc.src + "}"} {
				f := syntax.Parse("review.php", []byte(src), syntax.Options{})
				if len(f.Errors) > 0 {
					t.Fatalf("parse: %+v", f.Errors)
				}
				if got := e.Analyze(f); len(got) != tc.want {
					t.Fatalf("%s: got %d want %d: %+v", src, len(got), tc.want, got)
				}
			}
		})
	}
}
