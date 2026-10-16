package flow

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func TestExceptionEdgesRequirePotentialThrow(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		reached    bool
	}{
		{"literal", `try{1;'text';true;null;}catch(Throwable $e){caught();}`, false},
		{"fresh local", `try{$x=1;$x='text';$x;}catch(Throwable $e){caught();}`, false},
		{"known boolean", `$x=true;try{$x=1;}catch(Throwable $e){caught();}`, false},
		{"stored cast", `$x=(string)1;try{$x=1;}catch(Throwable $e){caught();}`, false},
		{"stored array cast", `$x=(array)1;try{$x=1;}catch(Throwable $e){caught();}`, true},
		{"calculated integer", `$x=1+2;try{$x=4;}catch(Throwable $e){caught();}`, false},
		{"known integer", `$x=1;try{$x=2;}catch(Throwable $e){caught();}`, false},
		{"loop object backedge", `$x=1;while($parameter){try{$x=2;}catch(Throwable $e){caught();}$x=new D;}`, true},
		{"known scalar", `$x='prior';try{$x=1;}catch(Throwable $e){caught();}`, false},
		{"boolean assignment", `$x=1;try{$y=!$x;$z=$x&&$y;$z;}catch(Throwable $e){caught();}`, false},
		{"literal arithmetic", `try{(1+2)*3;!false;}catch(Throwable $e){caught();}`, false},
		{"defined scalar read", `$parameter=1;try{$parameter;}catch(Throwable $e){caught();}`, false},
		{"unknown parameter read", `try{$parameter;}catch(Throwable $e){caught();}`, true},
		{"undefined local", `try{$missing;}catch(Throwable $e){caught();}`, true},
		{"unset local", `$x=1;unset($x);try{$x;}catch(Throwable $e){caught();}`, true},
		{"unknown prior", `$x=other();try{$x=1;}catch(Throwable $e){caught();}`, true},
		{"object destructor", `$x=new D;try{$x=1;}catch(Throwable $e){caught();}`, true},
		{"nested destructor", `$x=[[new D]];try{$x=1;}catch(Throwable $e){caught();}`, true},
		{"reference", `$x=1;$r=&$x;try{$x=2;}catch(Throwable $e){caught();}`, true},
		{"reference capture", `$x=1;$f=function()use(&$x){};try{$x=2;}catch(Throwable $e){caught();}`, true},
		{"global", `global $x;try{$x=1;}catch(Throwable $e){caught();}`, true},
		{"dynamic binding", `extract($parameter);try{$x=1;}catch(Throwable $e){caught();}`, true},
		{"switch comparison", `try{switch($parameter){case 'text':break;}}catch(Throwable $e){caught();}`, true},
		{"call condition", `try{if(other()){1;}}catch(Throwable $e){caught();}`, true},
		{"throw", `try{throw new Exception;}catch(Throwable $e){caught();}`, true},
		{"new", `try{new D;}catch(Throwable $e){caught();}`, true},
		{"include", `try{include 'other.php';}catch(Throwable $e){caught();}`, true},
		{"eval", `try{eval('other();');}catch(Throwable $e){caught();}`, true},
		{"property", `try{$parameter->p;}catch(Throwable $e){caught();}`, true},
		{"offset", `try{$parameter[0];}catch(Throwable $e){caught();}`, true},
		{"zero division", `try{1/0;}catch(Throwable $e){caught();}`, true},
		{"coercion", `try{$parameter+1;}catch(Throwable $e){caught();}`, true},
		{"output callback", `try{echo 'literal';}catch(Throwable $e){caught();}`, true},
		{"return destruction", `try{return 1;}catch(Throwable $e){caught();}`, true},
		{"unset destruction", `try{unset($parameter);}catch(Throwable $e){caught();}`, true},
		{"ticks", `register_tick_function(function(){throw new Exception;});try{declare(ticks=1){$x=1;}}catch(Throwable $e){caught();}`, true},
		{"this assignment", `try{$this=1;}catch(Throwable $e){caught();}`, true},
		{"GLOBALS assignment", `try{$GLOBALS=1;}catch(Throwable $e){caught();}`, true},
		{"read global", `try{$_SESSION;}catch(Throwable $e){caught();}`, true},
		{"read dynamic", `try{$$parameter;}catch(Throwable $e){caught();}`, true},
		{"nested assignment", `try{$x=($y=1);}catch(Throwable $e){caught();}`, true},
		{"read then write", `try{$x;$x=1;}catch(Throwable $e){caught();}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, `class D{function __destruct(){throw new Exception;}}function run($parameter){`+tc.body+`}`, nil)
			call := callsNamed(f, "caught")[0]
			var handler syntax.Node = call
			for handler != nil {
				if _, ok := handler.(*syntax.Catch); ok {
					break
				}
				handler = handler.Parent()
			}
			_, reached := e.scope(syntax.EnclosingVariableScope(call)).before[handler]
			if reached != tc.reached {
				t.Fatalf("catch reached=%v want %v", reached, tc.reached)
			}
		})
	}
}

func TestExceptionFramesKeepEntryAndPartialWrites(t *testing.T) {
	e, f := environment(t, `function run(){ $x=$_GET['before'];try{$x=$_POST['after'];other();}catch(Throwable $e){probe($x);}}`, nil)
	v := e.Value(argument(callsNamed(f, "probe")[0]))
	if !v.Complete || len(v.Sources) != 2 {
		t.Fatalf("entry or partial write lost: %+v", v)
	}
}

func TestPrunedExceptionFinallyPaths(t *testing.T) {
	for _, body := range []string{
		`try{1;}catch(Throwable $e){$x=$_GET['bad'];}finally{$x='final';}probe($x);`,
		`try{try{1;}finally{$x='final';}}catch(Throwable $e){$x=$_GET['bad'];}probe($x);`,
		`while($c){try{$x=1;break;}finally{$x='final';}}probe($x);`,
		`for($i=0;$i<2;$i++){try{$x=1;continue;}finally{$x='final';}}probe($x);`,
	} {
		e, f := environment(t, `function run($c){$x='start';`+body+`}`, nil)
		if e.Tainted(argument(callsNamed(f, "probe")[0]), HTML) {
			t.Fatalf("finally path changed: %s", body)
		}
	}
}

func TestExceptionProofWorkBudget(t *testing.T) {
	e, f := environment(t, `function run($p){try{`+strings.Repeat("!", 257)+`$p;}catch(Throwable $e){caught();}}`, nil)
	call := callsNamed(f, "caught")[0]
	var handler syntax.Node = call
	for handler != nil {
		if _, ok := handler.(*syntax.Catch); ok {
			break
		}
		handler = handler.Parent()
	}
	_, found := e.scope(syntax.EnclosingVariableScope(call)).before[handler]
	if !found {
		t.Fatal("proof budget exhaustion pruned catch")
	}
}

func BenchmarkExceptionEdges(b *testing.B) {
	for _, calls := range []bool{false, true} {
		var body strings.Builder
		for i := range 100 {
			if calls {
				fmt.Fprintf(&body, "other(%d);", i)
			} else {
				fmt.Fprintf(&body, "$v%d=%d;", i, i)
			}
		}
		src := `function run(){try{` + body.String() + `}catch(Throwable $e){sink($e);}}`
		b.Run(fmt.Sprint(calls), func(b *testing.B) {
			e, f := environment(b, src, nil)
			scope := f.Stmts[0]
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				New(f, e.types, nil).Calls(scope)
			}
		})
	}
}
