package flow

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func TestIndexedCallArgumentsPreserveOrderAndReachability(t *testing.T) {
	e, file := summarized(t, `function target($value){return $value;}function caller() {target(10);target(20);return;target(30);}`)
	calls := callsNamed(file, "target")
	for i, call := range calls {
		values, known := e.BoundArguments(call)
		if i == 2 {
			if known {
				t.Fatal("unreached call acquired arguments")
			}
			continue
		}
		if !known || len(values) != 1 {
			t.Fatal("missing call arguments")
		}
		if value, ok := literalInteger(values[0].Expr); !ok || value != int64((i+1)*10) {
			t.Fatal("call arguments crossed sites")
		}
	}
	history := e.Calls(syntax.EnclosingVariableScope(calls[0]))
	if len(history) != 2 || history[0].Node != calls[0] || history[1].Node != calls[1] {
		t.Fatal("ordered call history changed")
	}
}

func TestIndexedCallsAfterNativeReplay(t *testing.T) {
	e, file := summarized(t, `function target($value){return $value;}function caller(){$callback=strtolower(...);$callback('A');target(10);$callback('B');target(20);}`)
	calls := callsNamed(file, "target")
	for i, call := range calls {
		values, known := e.BoundArguments(call)
		if !known || len(values) != 1 {
			t.Fatal("replay dropped indexed call")
		}
		if value, ok := literalInteger(values[0].Expr); !ok || value != int64((i+1)*10) {
			t.Fatal("replay shifted call index")
		}
	}
	for _, call := range e.Calls(syntax.EnclosingVariableScope(calls[0])) {
		if call.Name == "strtolower" {
			t.Fatal("synthetic native replay leaked into history")
		}
	}
}

func BenchmarkIndexedArgumentLookup(b *testing.B) {
	for _, count := range []int{128, 1024} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			var source strings.Builder
			source.WriteString(`function target($value){return $value;}function caller(){`)
			for i := range count {
				fmt.Fprintf(&source, "target(%d);", i)
			}
			source.WriteString("}")
			e, file := summarized(b, source.String())
			calls := callsNamed(file, "target")
			last := calls[len(calls)-1]
			e.BoundArguments(last)
			b.ReportAllocs()
			for b.Loop() {
				if _, known := e.BoundArguments(last); !known {
					b.Fatal("lost call")
				}
			}
		})
	}
}

func TestIndexedWrapperWithNestedNativeArgument(t *testing.T) {
	definition, defs := environment(t, `function net($h,$x){curl_setopt($h,CURLOPT_URL,$x);}`, nil)
	snapshot := NewSnapshot(Extract(defs, definition.types.Index, definition.types.PHP, nil, Options{}))
	e, file := environment(t, `net(curl_init(),$_GET['x']);`, snapshot)
	e.types.Index = definition.types.Index
	at := callsNamed(file, "net")[0]
	args, known := e.BoundArguments(at)
	if !known || len(args) != 2 || !tainted(args[1]) {
		t.Fatalf("args=%+v known=%v", args, known)
	}
	if sinks := e.Sinks(at); len(sinks) != 1 {
		t.Fatalf("sinks=%+v", sinks)
	}
}
