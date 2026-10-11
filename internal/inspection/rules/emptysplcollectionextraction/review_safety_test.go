package emptysplcollectionextraction

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, Only: []string{"EmptySplCollectionExtraction"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		// Loop bodies replay: a later insertion reaches the call on the next pass.
		{"foreach later push", `$s=new SplStack();foreach($t as $v){if($v){$s->pop();}else{$s->push($v);}}`, 0},
		{"while later insert", `$h=new SplMinHeap();while($c){$h->extract();$h->insert(1);}`, 0},
		{"do-while later enqueue", `$q=new SplQueue();do{$q->dequeue();$q->enqueue(1);}while($c);`, 0},
		{"for later unshift", `$q=new SplQueue();for(;;){$q->shift();$q->unshift(1);}`, 0},
		{"nested loop later push", `$q=new SplQueue();while($a){while($b){$q->pop();}$q->push(1);}`, 0},
		{"loop later escape", `$q=new SplQueue();while($c){$q->pop();fill($q);}`, 0},
		{"loop later unmodeled", `$q=new SplQueue();while($c){$q->pop();$q->add(0,1);}`, 0},
		{"loop removal only", `$q=new SplQueue();while($c){$q->pop();}`, 1},
		{"loop later removal", `$q=new SplQueue();while($c){$q->top();$q->pop();}`, 2},
		{"loop earlier conditional removal", `$q=new SplQueue();while($c){if($d){$q->pop();}$q->top();}`, 1},
		{"loop rebuilds collection", `while($c){$q=new SplQueue();$q->pop();$q->push(1);}`, 1},
		{"loop after call", `$q=new SplQueue();$q->pop();while($c){$q->push(1);}`, 1},
		{"label before call", `$q=new SplQueue();again:if($n++){$q->pop();}$q->push(1);if($n<2)goto again;`, 0},
		{"label after call", `$q=new SplQueue();$q->pop();again:$q->push(1);if($n++<2)goto again;`, 1},

		// Guards outside a plain if statement.
		{"ternary else", `$q=new SplQueue();$x=$q->isEmpty()?null:$q->dequeue();`, 0},
		{"ternary then", `$q=new SplQueue();$x=$q->count()>0?$q->bottom():null;`, 0},
		{"short ternary", `$q=new SplQueue();$x=$q->isEmpty()?:$q->dequeue();`, 0},
		{"ternary wrong arm", `$q=new SplQueue();$x=$q->isEmpty()?$q->dequeue():null;`, 1},
		{"ternary unrelated", `$q=new SplQueue();$x=$c?$q->dequeue():null;`, 1},
		{"or guard", `$q=new SplQueue();$q->isEmpty()||$q->dequeue();`, 0},
		{"word or guard", `$q=new SplQueue();$q->isEmpty() or $q->dequeue();`, 0},
		{"and guard", `$q=new SplQueue();!$q->isEmpty()&&$q->dequeue();`, 0},
		{"word and guard", `$q=new SplQueue();!$q->isEmpty() and $q->dequeue();`, 0},
		{"and wrong polarity", `$q=new SplQueue();$q->isEmpty()&&$q->dequeue();`, 1},
		{"or wrong polarity", `$q=new SplQueue();!$q->isEmpty()||$q->dequeue();`, 1},
		{"left operand", `$q=new SplQueue();$x=$q->dequeue()&&!$q->isEmpty();`, 1},
		{"other operator", `$q=new SplQueue();$x=!$q->isEmpty()^$q->dequeue();`, 1},
		{"nested in and", `$q=new SplQueue();if(!$q->isEmpty()&&$q->top()===1){echo 1;}`, 0},
		{"conjunction condition", `$q=new SplQueue();if($c&&!$q->isEmpty()){$q->dequeue();}`, 0},
		{"conjunction left", `$q=new SplQueue();if(!$q->isEmpty() and $c){$q->dequeue();}`, 0},
		{"conjunction refuted", `$q=new SplQueue();if($c&&!$q->isEmpty()){}else{$q->dequeue();}`, 1},
		{"disjunction refuted", `$q=new SplQueue();if($c||$q->isEmpty()){return;}$q->dequeue();`, 0},
		{"disjunction left refuted", `$q=new SplQueue();if($q->isEmpty() or $c){}else{$q->dequeue();}`, 0},
		{"disjunction held", `$q=new SplQueue();if($c||!$q->isEmpty()){$q->dequeue();}`, 1},
		{"while guard", `$q=new SplQueue();while(!$q->isEmpty()){$q->dequeue();}`, 0},
		{"while count guard", `$q=new SplQueue();while($q->count()>0){$q->pop();}`, 0},
		{"while condition", `$q=new SplQueue();while($q->dequeue()){}`, 1},
		{"for guard", `$q=new SplQueue();for(;!$q->isEmpty();){$q->dequeue();}`, 0},
		{"for last condition", `$q=new SplQueue();for(;$c,!$q->isEmpty();){$q->dequeue();}`, 0},
		{"for without condition", `$q=new SplQueue();for(;;){$q->dequeue();}`, 1},
		{"for step", `$q=new SplQueue();for(;!$q->isEmpty();$q->dequeue()){}`, 1},
		{"elseif own guard", `$q=new SplQueue();if($c){}elseif(!$q->isEmpty()){$q->dequeue();}`, 0},
		{"elseif after empty", `$q=new SplQueue();if($q->isEmpty()){}elseif($c){$q->dequeue();}`, 0},
		{"elseif chain", `$q=new SplQueue();if($a){}elseif($q->isEmpty()){}elseif($b){$q->dequeue();}`, 0},
		{"else after elseif", `$q=new SplQueue();if($a){}elseif($q->isEmpty()){}else{$q->dequeue();}`, 0},
		{"elseif unrelated", `$q=new SplQueue();if($a){}elseif($b){$q->dequeue();}`, 1},
		{"later elseif", `$q=new SplQueue();if($a){}elseif($b){$q->dequeue();}elseif($q->isEmpty()){}`, 1},
		{"else unrelated", `$q=new SplQueue();if($a){}elseif($b){}else{$q->dequeue();}`, 1},
		{"if condition", `$q=new SplQueue();if($q->dequeue()){}`, 1},
		{"elseif condition", `$q=new SplQueue();if($a){}elseif($q->dequeue()){}`, 1},

		// Any use besides a method receiver may hand the collection away.
		{"property store", `$q=new SplQueue();$o->q=$q;$o->fill();$q->dequeue();`, 0},
		{"array literal", `$q=new SplQueue();$h=['q'=>$q];$h['q']->enqueue(1);$q->dequeue();`, 0},
		{"plain alias", `$a=new SplQueue();$b=$a;$b->enqueue(1);$a->dequeue();`, 0},
		{"reference alias", `$q=new SplQueue();$b=&$q;$b->enqueue(1);$q->dequeue();`, 0},
		{"method argument", `$q=new SplQueue();$q->enqueue($q);$q->dequeue();$q->dequeue();`, 0},
		{"returned by closure call", `$q=new SplQueue();$f=[$q,'push'];$f(1);$q->pop();`, 0},
		{"array append", `$q=new SplQueue();$q[]=1;$q->dequeue();`, 0},
		{"unrelated reference", `$q=new SplQueue();$a=&$b;$q->dequeue();`, 1},
		{"unrelated uses", `$q=new SplQueue();$o->q=$p;$h=[$p];$p->enqueue(1);$q->dequeue();`, 1},
		{"direct construction", `(new SplQueue())->dequeue();`, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte("<?php "+tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
