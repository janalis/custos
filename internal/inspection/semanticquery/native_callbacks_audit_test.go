package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

func TestNativeCallbackLiteralReachability(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{"function f(){if(false){probe();}}", false},
		{"function f(){if(true){}else{probe();}}", false},
		{"function f(){return;probe();}", false},
		{"function f(){if($unknown){probe();}}", true},
		{"function f(){if(true){probe();}}", true},
		{"function f(){if(false){}else{probe();}}", true},
		{"probe();", true},
		{"function f(){if(0){probe();}}", false},
		{"function f(){if(1){}else{probe();}}", false},
		{"function f(){if(2){probe();}}", true},
		{"function f(){if(0.0){probe();}}", true},
		{"function f(){if(0 && $x){probe();}}", false},
		{"function f(){if((1 or $x)){}else{probe();}}", false},
		{"function f(){if(1 && $x){probe();}}", true},
		{"function f(){if($x . 1){probe();}}", true},
		{"function f(){return 0 && probe();}", false},
		{"function f(){return false || probe();}", true},
		{"function f(){return probe() && 0;}", true},
		{"function f(){return 1 + probe();}", true},
	} {
		t.Run(tc.src, func(t *testing.T) {
			f := syntax.Parse("reach.php", []byte("<?php "+tc.src), syntax.Options{})
			checked := false
			syntax.InspectFile(f, func(n syntax.Node) bool {
				if c, ok := n.(*syntax.FuncCall); ok {
					checked = true
					if got := nativeCallbackReachable(c); got != tc.want {
						t.Fatalf("got %v, want %v", got, tc.want)
					}
				}
				return true
			})
			if !checked {
				t.Fatal("missing probe")
			}
		})
	}
}

func BenchmarkNativeCallbackReachable(b *testing.B) {
	f := syntax.Parse("reach.php", []byte("<?php function f($x){if($x){probe();}}"), syntax.Options{})
	var call *syntax.FuncCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			call = c
		}
		return true
	})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		nativeCallbackReachable(call)
	}
}

func TestAuditCurlMethodEscape(t *testing.T) {
	probeNative(t, "class Holder{function retain($h){}} $owner=new Holder();$h=curl_init('https://example.invalid');$owner->retain($h);probe($h);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if NativeCurlDefault(ctx, c, CallArgument(c.Args, 0, ""), "CURLOPT_RETURNTRANSFER") {
			t.Fatal("method escape preserved default proof")
		}
	})
}

func TestAuditFailureGuardsTrackReachingValue(t *testing.T) {
	for _, src := range []string{
		"$v=preg_replace('/x/','y','x');if($v!==null){$v=preg_replace('/x/','y','x');probe($v);}",
		"if(preg_replace('/x/','y','x')!==null){probe(preg_replace('/x/','y','x'));}",
	} {
		t.Run(src, func(t *testing.T) {
			probeNative(t, src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				if NativeFailureExcluded(ctx, CallArgument(c.Args, 0, ""), types.Unknown, "null") {
					t.Fatal("old producer guard applied to new result")
				}
			})
		})
	}
}
