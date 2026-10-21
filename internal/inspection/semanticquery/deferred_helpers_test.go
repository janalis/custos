package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestConstructionHonorsClassAvailability(t *testing.T) {
	for _, version := range []phpversion.Version{phpversion.PHP74, phpversion.PHP80} {
		seen := false
		p := nativeProbe{check: func(ctx *analysis.Context, call *syntax.FuncCall) {
			seen = true
			got := NativeConstruction(ctx, CallArgument(call.Args, 0, ""), "WeakMap")
			if (got != nil) != (version >= phpversion.PHP80) {
				t.Fatalf("PHP %s: construction=%v", version, got)
			}
		}}
		e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{PHP: version, Only: []string{p.ID()}})
		if err != nil {
			t.Fatal(err)
		}
		e.Analyze(syntax.Parse("available.php", []byte(`<?php $map=new WeakMap; probe($map);`), syntax.Options{}))
		if !seen {
			t.Fatal("probe did not run")
		}
	}
}

func TestNamedCallbacksNeedAvailableDeclaration(t *testing.T) {
	for _, tc := range []struct {
		code  string
		known bool
	}{
		{`function handler($value){return $value;} probe('handler');`, true},
		{`namespace App; function handler($value){return $value;} probe('App\handler');`, true},
		{`if(false){function handler($value){return $value;}} probe('handler');`, false},
		{`function outer(){function handler($value){return $value;}} probe('handler');`, false},
		{`if($condition){function handler($value){return $value;}}else{function handler($value){return 1;}} probe('handler');`, false},
	} {
		t.Run(tc.code, func(t *testing.T) {
			probeNative(t, tc.code, func(ctx *analysis.Context, call *syntax.FuncCall) {
				_, body := NativeCallback(ctx, CallArgument(call.Args, 0, ""))
				if (body != nil) != tc.known {
					t.Fatalf("callback resolved=%v, want %v", body != nil, tc.known)
				}
			})
		})
	}
}

func TestMemoryStreamEffectiveMode(t *testing.T) {
	for _, tc := range []struct{ mode, want string }{
		{"r", "r"}, {"rb", "r"}, {"rw", "w+"}, {"r+", "w+"}, {"w", "w+"}, {"ra", "a+"},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			probeNative(t, `$stream=fopen('php://memory','`+tc.mode+`');if($stream===false)return;probe($stream);`, func(ctx *analysis.Context, call *syntax.FuncCall) {
				mode, known := NativeStreamMode(ctx, CallArgument(call.Args, 0, ""), call)
				if !known || mode != tc.want {
					t.Fatalf("mode=%q known=%v, want %q", mode, known, tc.want)
				}
			})
		})
	}
}

func TestSQLVendorCommentsDoNotEstablishGenericCounts(t *testing.T) {
	for _, sql := range []string{`SELECT /*! ? */ ?`, `SELECT /*M! ? */ ?`, `SELECT /*!50700 ? */ ?`, `SELECT 1--? + ?`} {
		if facts, known := ScanSQL(sql); known {
			t.Fatalf("dialect-sensitive SQL claimed complete: %s %+v", sql, facts)
		}
	}
}

func TestOpaqueHandleOwnership(t *testing.T) {
	for _, tc := range []struct {
		store string
		owned bool
	}{
		{``, true},
		{`$other=curl_init();`, true},
		{`$copy=$handle;`, false},
		{`$copy=&$handle;`, false},
		{`$handle=&$copy;`, false},
		{`$saved=[$handle];`, false},
		{`$saved=['nested'=>[$handle]];`, false},
		{`$object->saved=$handle;`, false},
		{`State::$saved=$handle;`, false},
		{`$callback=function()use($handle){};`, false},
		{`$callback=fn()=> $handle;`, false},
		{`$callback=fn($handle)=> $handle;`, true},
		{`$callback=function()use(&$handle){};`, false},
	} {
		t.Run(tc.store, func(t *testing.T) {
			probeNative(t, `$handle=curl_init();`+tc.store+`probe($handle);`, func(ctx *analysis.Context, call *syntax.FuncCall) {
				handle := CallArgument(call.Args, 0, "")
				if got := NativeHandleUnescaped(ctx, handle, call); got != tc.owned {
					t.Fatalf("exclusive ownership=%v, want %v", got, tc.owned)
				}
				if !tc.owned && NativeCurlDefault(ctx, call, handle, "CURLOPT_HEADER") {
					t.Fatal("escaped handle retained default configuration proof")
				}
			})
		})
	}
}

func TestHandleOwnershipBudgetIsConservative(t *testing.T) {
	probeNative(t, `$handle=curl_init();$capture=fn()=>[`+strings.Repeat(`1,`, 5000)+`];probe($handle);`, func(ctx *analysis.Context, call *syntax.FuncCall) {
		if NativeHandleUnescaped(ctx, CallArgument(call.Args, 0, ""), call) {
			t.Fatal("exhausted capture scan established exclusive ownership")
		}
	})
}

func TestShadowedCurlOptionDoesNotEstablishDefault(t *testing.T) {
	probeNative(t, `namespace App;const CURLOPT_HEADER=99;$handle=curl_init();curl_setopt($handle,CURLOPT_HEADER,false);probe($handle);`, func(ctx *analysis.Context, call *syntax.FuncCall) {
		if NativeCurlDefault(ctx, call, CallArgument(call.Args, 0, ""), "CURLOPT_HEADER") {
			t.Fatal("project constant established a builtin option default")
		}
	})
}

func TestCurlDefaultsRequireCurlAllocation(t *testing.T) {
	for _, source := range []string{
		`$handle=tmpfile();if($handle===false)return;probe($handle);`,
		`$handle=new stdClass();probe($handle);`,
	} {
		probeNative(t, source, func(ctx *analysis.Context, call *syntax.FuncCall) {
			if NativeCurlDefault(ctx, call, CallArgument(call.Args, 0, ""), "CURLOPT_HEADER") {
				t.Fatal("another allocation established cURL defaults")
			}
		})
	}
}
