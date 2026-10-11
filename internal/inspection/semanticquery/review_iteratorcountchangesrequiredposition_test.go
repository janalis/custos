package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// reviewIteratorCountProbe dispatches on method calls only, as the rule package does.
type reviewIteratorCountProbe struct{ nativeCheckProbe }

func (reviewIteratorCountProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall}
}

func TestReviewIteratorCountChangesRequiredPosition(t *testing.T) {
	const counted = "$it=new ArrayIterator([7,11]);$n=iterator_count($it);"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"read after count", counted + "echo $it->current();", 1},
		{"key after count", counted + "echo $it->key();", 1},
		{"both reads", counted + "echo $it->key(),$it->current();", 2},
		{"in function", "function f(){" + counted + "return $it->current();}", 1},
		{"parenthesized receiver", counted + "echo ($it)->current();", 1},
		{"valid keeps position", counted + "if($it->valid()){}echo $it->current();", 1},
		{"next keeps position", counted + "$it->next();echo $it->current();", 1},
		{"count method keeps position", counted + "echo $it->count(),$it->current();", 1},
		{"rewind before last count", counted + "$it->rewind();iterator_count($it);echo $it->current();", 1},
		{"seek before last count", counted + "$it->seek(1);iterator_count($it);echo $it->current();", 1},
		{"later rewind in loop body", counted + "while($more){echo $it->current();$it->rewind();}", 1},
		{"unrelated loop", counted + "foreach($rows as $row){echo $it->current();}", 1},
		{"other function scope", counted + "function g(){$it=1;}echo $it->current();", 1},
		{"foreach over iterator", counted + "foreach($it as $v){echo $it->key(),$v;}", 0},
		{"foreach with key", counted + "foreach($it as $k=>$v){echo $it->current();}", 0},
		{"for init rewind", counted + "for($it->rewind();$it->valid();$it->next()){echo $it->key(),$it->current();}", 0},
		{"for init seek", counted + "for($it->seek(0);$it->valid();$it->next()){echo $it->current();}", 0},
		{"conditional rewind", counted + "if($b){$it->rewind();}echo $it->current();", 0},
		{"rewind", counted + "$it->rewind();echo $it->current();", 0},
		{"seek", counted + "$it->seek(0);echo $it->current();", 0},
		{"wrapper iterator", counted + "$l=new LimitIterator($it,0,1);foreach($l as $x){echo $it->current();}", 0},
		{"offset write", counted + "$it[5]=1;echo $it->current();", 0},
		{"alias before count", "$it=new ArrayIterator([7,11]);$a=$it;iterator_count($it);$a->rewind();echo $it->current();", 0},
		{"alias after count", counted + "$a=$it;$a->rewind();echo $it->current();", 0},
		{"closure capture", counted + "$f=function()use($it){$it->rewind();};$f();echo $it->current();", 0},
		{"arrow function capture", counted + "$f=fn()=>$it->rewind();$f();echo $it->current();", 0},
		{"unknown call", counted + "restore($it);echo $it->current();", 0},
		{"unknown method", counted + "$it->uasort('strcmp');echo $it->current();", 0},
		{"reassigned", counted + "$it=new ArrayIterator([3]);echo $it->current();", 0},
		{"empty iterator", "$it=new ArrayIterator([]);iterator_count($it);echo $it->key();", 0},
		{"direct construction", "echo (new ArrayIterator([1]))->current();", 0},
		{"variable variable", "$n='it';" + counted + "echo $$n->current();", 0},
		{"conditional count", "$it=new ArrayIterator([7,11]);if($b){iterator_count($it);}echo $it->current();", 0},
		{"no count", "$it=new ArrayIterator([7,11]);echo $it->current();", 0},
		{"count argument missing", "$it=new ArrayIterator([7,11]);iterator_count();echo $it->current();", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := reviewIteratorCountProbe{nativeCheckProbe{"IteratorCountChangesRequiredPosition", CheckIteratorCountChangesRequiredPosition}}
			e, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{"IteratorCountChangesRequiredPosition"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
