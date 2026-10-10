package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeCurlExplicitOption(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"$h=curl_init();curl_setopt($h,CURLOPT_HEADER,true);probe($h);", "true"},
		{"$h=curl_init();curl_setopt($h,CURLOPT_HEADER,true);curl_setopt($h,CURLOPT_HEADER,false);probe($h);", "false"},
		{"$h=curl_init();curl_setopt($h,CURLOPT_RETURNTRANSFER,true);probe($h);", ""},
		{"function transfer(){$h=curl_init();curl_setopt($h,CURLOPT_HEADER,true);probe($h);}", "true"},
		{"$h=curl_init();probe($h);", ""},
		{"probe($unknown);", ""},
		{"$h=fopen('x','r');probe($h);", ""},
		{"$h=curl_init();curl_setopt($h,CURLOPT_HEADER,true);retain($h);probe($h);", ""},
		{"$h=curl_init();curl_setopt_array($h,[CURLOPT_HEADER=>true]);probe($h);", ""},
		{"$h=curl_init();if($test){curl_setopt($h,CURLOPT_HEADER,true);}probe($h);", ""},
		{"$h=curl_init();curl_setopt($h,$option,true);probe($h);", ""},
		{"namespace Local; const CURLOPT_HEADER=1;$h=curl_init();curl_setopt($h,CURLOPT_HEADER,true);probe($h);", ""},
		{"const CUSTOM=CURLOPT_HEADER;$h=curl_init();curl_setopt($h,CURLOPT_HEADER,true);curl_setopt($h,CUSTOM,false);probe($h);", ""},
		{"const CUSTOM=1;$h=curl_init();curl_setopt($h,CUSTOM,true);probe($h);", ""},
		{"$h=curl_init();curl_setopt($h,CURLOPT_HEADER);probe($h);", ""},
		{"$h=curl_init();curl_setopt($h,CURLOPT_HEADER,true);probe($h);curl_setopt($h,CURLOPT_HEADER,false);", "true"},
		{"$h=curl_init();$other=curl_init();curl_setopt($other,CURLOPT_HEADER,true);probe($h);", ""},
	} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				value, known := NativeCurlOption(ctx, c, CallArgument(c.Args, 0, "value"), "CURLOPT_HEADER")
				if got := ctx.Text(value); got != tc.want || known != (tc.want != "") {
					t.Fatalf("got %q %v want %q", got, known, tc.want)
				}
				NativeCurlOption(ctx, c, CallArgument(c.Args, 0, "value"), "CURLOPT_HEADER")
			})
		})
	}
}

func BenchmarkNativeCurlExplicitOption(b *testing.B) {
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		if NativeBuiltin(ctx, c, "curl_exec") {
			NativeCurlOption(ctx, c, CallArgument(c.Args, 0, "handle"), "CURLOPT_HEADER")
		}
	}}
	engine, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	source := []byte("<?php $h=curl_init();curl_setopt($h,CURLOPT_HEADER,true);curl_exec($h);")
	b.ReportAllocs()
	for b.Loop() {
		engine.Analyze(syntax.Parse("bench.php", source, syntax.Options{}))
	}
}
