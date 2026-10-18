package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestConditionalHistoryIsRetained(t *testing.T) {
	source := `$handle=curl_init();curl_setopt($handle,CURLOPT_HEADER,false);if($flag){curl_setopt($handle,CURLOPT_HEADER,true);}probe($handle);`
	probeNative(t, source, func(ctx *analysis.Context, at *syntax.FuncCall) {
		history := NativeCallHistory(ctx, at, CallArgument(at.Args, 0, ""), "curl_setopt")
		if !history.Complete || len(history.Events) != 2 || !history.Events[0].Must || history.Events[1].Must {
			t.Fatalf("history=%+v", history)
		}
		if _, known := history.Certain(); known {
			t.Fatal("conditional overwrite became an unconditional history")
		}
		if history.Events[0].Node.Span().Start >= history.Events[1].Node.Span().Start {
			t.Fatal("history is not ordered")
		}
	})
}

func TestKnownAndUnknownOperationHistories(t *testing.T) {
	for _, tc := range []struct {
		code  string
		count int
		known bool
	}{
		{`$handle=curl_init();probe($handle);`, 0, true},
		{`$handle=curl_init();$alias=$handle;curl_setopt($alias,CURLOPT_HEADER,true);probe($handle);`, 1, true},
		{`$handle=curl_init();$holder->value=$handle;probe($handle);`, 0, false},
		{`$handle=curl_init();$alias=$handle;$callback=fn()=>curl_setopt($alias,CURLOPT_HEADER,true);probe($handle);`, 0, false},
		{`$handle=curl_init();$callback=function()use($handle){};probe($handle);`, 0, false},
		{`$handle=curl_init();curl_setopt($handle,CURLOPT_HEADER,false);curl_setopt($handle,CURLOPT_HEADER,true);probe($handle);`, 2, true},
		{`$handle=curl_init();$other=curl_init();if($flag){curl_setopt($other,CURLOPT_HEADER,true);}probe($handle);`, 0, true},
		{`$handle=curl_init();probe($handle);curl_setopt($handle,CURLOPT_HEADER,true);`, 0, true},
		{`$handle=curl_init();while($flag){probe($handle);curl_setopt($handle,CURLOPT_HEADER,true);}`, 0, false},
		{`$handle=curl_init();curl_setopt(...$arguments);probe($handle);`, 0, false},
		{`probe($unknown);`, 0, false},
		{`function configure($value){curl_setopt($value,CURLOPT_HEADER,true);}$handle=curl_init();configure($handle);probe($handle);`, 0, false},
		{`$handle=curl_init();curl_setopt($unknown,CURLOPT_HEADER,true);probe($handle);`, 0, false},
	} {
		t.Run(tc.code, func(t *testing.T) {
			probeNative(t, tc.code, func(ctx *analysis.Context, at *syntax.FuncCall) {
				history := NativeCallHistory(ctx, at, CallArgument(at.Args, 0, ""), "curl_setopt")
				calls, known := history.Certain()
				if known != tc.known || len(calls) != tc.count {
					t.Fatalf("known=%v calls=%d history=%+v", known, len(calls), history)
				}
			})
		})
	}
}

func TestHistoryWorkIsBounded(t *testing.T) {
	source := `$handle=curl_init();` + strings.Repeat(`strlen('fixed');`, maxHistoryWork+1) + `probe($handle);`
	probeNative(t, source, func(ctx *analysis.Context, at *syntax.FuncCall) {
		if history := NativeCallHistory(ctx, at, CallArgument(at.Args, 0, ""), "curl_setopt"); history.Complete {
			t.Fatal("exhausted history established absence")
		}
	})
}

func BenchmarkOperationHistory(b *testing.B) {
	p := nativeProbe{check: func(ctx *analysis.Context, at *syntax.FuncCall) {
		name, ok := at.Name.(*syntax.Name)
		if !ok || name.Value != "probe" {
			return
		}
		receiver := CallArgument(at.Args, 0, "")
		ctx.Flow().Value(receiver)
		b.ReportAllocs()
		for b.Loop() {
			NativeCallHistory(ctx, at, receiver, "curl_setopt").Certain()
		}
	}}
	engine, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	engine.Analyze(syntax.Parse("history.php", []byte(`<?php $handle=curl_init();`+strings.Repeat(`curl_setopt($handle,CURLOPT_HEADER,false);`, 10)+`probe($handle);`), syntax.Options{}))
}

func TestHistoryInternalWorkLimits(t *testing.T) {
	for _, tc := range []struct {
		source string
		budget int
	}{
		{`$handle=curl_init();curl_setopt($handle,CURLOPT_HEADER,false);probe($handle);`, 1},
		{`$handle=curl_init();probe($handle);if($flag){curl_setopt($handle,CURLOPT_HEADER,true);}`, 3},
	} {
		probeNative(t, tc.source, func(ctx *analysis.Context, at *syntax.FuncCall) {
			receiver := CallArgument(at.Args, 0, "")
			if value := ctx.Flow().Value(receiver); !value.Complete || value.Identity == 0 {
				t.Fatal("work-limit control has no identity proof")
			}
			if history := nativeCallHistory(ctx, at, receiver, tc.budget, "curl_setopt"); history.Complete {
				t.Fatal("work budget was ignored")
			}
		})
	}
}

func TestPartialHistoryRetainsProvenSuffix(t *testing.T) {
	probeNative(t, `function configure($value){curl_setopt($value,CURLOPT_HEADER,true);}$handle=curl_init();configure($handle);curl_setopt($handle,CURLOPT_HEADER,false);probe($handle);`, func(ctx *analysis.Context, at *syntax.FuncCall) {
		history := NativeCallHistory(ctx, at, CallArgument(at.Args, 0, ""), "curl_setopt")
		if history.Complete || len(history.Events) != 1 || !history.Events[0].Must {
			t.Fatalf("history=%+v", history)
		}
		if _, known := history.Certain(); known {
			t.Fatal("partial suffix became a complete initial-state proof")
		}
	})
}

func TestLoopAllocationResetsOperationHistory(t *testing.T) {
	for _, loop := range []string{
		`while($again){$handle=curl_init();probe($handle);curl_setopt($handle,CURLOPT_HEADER,true);}`,
		`do{$handle=curl_init();probe($handle);curl_setopt($handle,CURLOPT_HEADER,true);}while($again);`,
		`for(;$again;){$handle=curl_init();probe($handle);curl_setopt($handle,CURLOPT_HEADER,true);}`,
		`foreach($items as $item){$handle=curl_init();probe($handle);curl_setopt($handle,CURLOPT_HEADER,true);}`,
	} {
		t.Run(loop, func(t *testing.T) {
			probeNative(t, `function example($again,$items){`+loop+`}`, func(ctx *analysis.Context, at *syntax.FuncCall) {
				calls, known := NativeCallHistory(ctx, at, CallArgument(at.Args, 0, ""), "curl_setopt").Certain()
				if !known || len(calls) != 0 {
					t.Fatal("previous iteration's object affected a fresh allocation")
				}
			})
		})
	}
}

func TestLoopInitializerDoesNotResetEveryTurn(t *testing.T) {
	probeNative(t, `for($handle=curl_init();$again;){probe($handle);curl_setopt($handle,CURLOPT_HEADER,true);}`, func(ctx *analysis.Context, at *syntax.FuncCall) {
		if history := NativeCallHistory(ctx, at, CallArgument(at.Args, 0, ""), "curl_setopt"); history.Complete {
			t.Fatal("loop initializer was treated as a per-turn allocation")
		}
	})
}

func TestCoalescingAssignmentCallIsConditional(t *testing.T) {
	probeNative(t, `$handle=curl_init();$done??=curl_setopt($handle,CURLOPT_HEADER,true);probe($handle);`, func(ctx *analysis.Context, at *syntax.FuncCall) {
		history := NativeCallHistory(ctx, at, CallArgument(at.Args, 0, ""), "curl_setopt")
		if len(history.Events) != 1 || history.Events[0].Must {
			t.Fatalf("coalescing assignment became unconditional: %+v", history)
		}
	})
}

func TestHistoryFreshHandleWithJoinedInput(t *testing.T) {
	probeNative(t, `$n='https://api.example.com/?q='.$_GET['q'];try{run();}catch(Exception $n){}$handle=curl_init($n);probe($handle);`, func(ctx *analysis.Context, at *syntax.FuncCall) {
		receiver := CallArgument(at.Args, 0, "")
		value := ctx.Flow().Value(receiver)
		history := NativeCallHistory(ctx, at, receiver, "curl_setopt")
		if !history.Complete {
			t.Fatalf("value=%+v history=%+v", value, history)
		}
	})
}

func TestHistoryAllocationRestartRequiresDirectAssignment(t *testing.T) {
	for _, tc := range []struct {
		code string
		want bool
	}{
		{`while($flag){$handle=((curl_init()));probe($handle);}`, true},
		{`while($flag){consume(curl_init());probe($handle);}`, false},
		{`while($flag){$handle+=curl_init();probe($handle);}`, false},
		{`while($flag){$other=curl_init();probe($handle);}`, false},
	} {
		t.Run(tc.code, func(t *testing.T) {
			file := syntax.Parse("restart.php", []byte("<?php "+tc.code), syntax.Options{})
			var origin syntax.Expr
			var receiver syntax.Expr
			var loop, at syntax.Node
			syntax.InspectFile(file, func(node syntax.Node) bool {
				if _, ok := node.(*syntax.While); ok {
					loop = node
				}
				if call, ok := node.(*syntax.FuncCall); ok {
					if name, ok := call.Name.(*syntax.Name); ok {
						if name.Value == "curl_init" {
							origin = call
						}
						if name.Value == "probe" {
							at = call
							receiver = CallArgument(call.Args, 0, "")
						}
					}
				}
				return true
			})
			if origin == nil || receiver == nil || loop == nil || at == nil {
				t.Fatal("invalid restart probe")
			}
			if got := historyAllocationRestarts(receiver, origin, origin.Span().Start+1, loop, at); got != tc.want {
				t.Fatalf("restart=%v want %v", got, tc.want)
			}
			if historyAllocationRestarts(origin, origin, origin.Span().Start+1, loop, at) {
				t.Fatal("nonvariable receiver restarted")
			}
		})
	}
}

func TestHistoryMissingReceiverDoesNotCreateAnOperation(t *testing.T) {
	probeNative(t, `$handle=curl_init();curl_setopt();probe($handle);`, func(ctx *analysis.Context, at *syntax.FuncCall) {
		calls, known := NativeCallHistory(ctx, at, CallArgument(at.Args, 0, ""), "curl_setopt").Certain()
		if !known || len(calls) != 0 {
			t.Fatalf("known=%v calls=%v", known, calls)
		}
	})
}

func TestNestedOperationsFollowEvaluationOrder(t *testing.T) {
	probeNative(t, `$handle=curl_init();curl_setopt($handle,CURLOPT_HEADER,curl_setopt($handle,CURLOPT_HEADER,false));probe($handle);`, func(ctx *analysis.Context, at *syntax.FuncCall) {
		calls, known := NativeCallHistory(ctx, at, CallArgument(at.Args, 0, ""), "curl_setopt").Certain()
		if !known || len(calls) != 2 || !calls[1].Span().Contains(calls[0].Span()) {
			t.Fatalf("known=%v calls=%v", known, calls)
		}
	})
}

func TestNonNativeDocumentLoaderCannotEstablishCertainHistory(t *testing.T) {
	probeNative(t, `final class CustomDocument extends DOMDocument{function loadXML(string $source,int $options=0):bool{return false;}}$document=new CustomDocument();$document->loadXML('<root/>');probe($document);`, func(ctx *analysis.Context, at *syntax.FuncCall) {
		if history := NativeCallHistory(ctx, at, CallArgument(at.Args, 0, ""), "loadXML"); history.Complete {
			t.Fatal("custom loader established a native document history")
		}
	})
}
