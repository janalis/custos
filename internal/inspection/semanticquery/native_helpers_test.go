package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

type nativeProbe struct {
	check func(*analysis.Context, *syntax.FuncCall)
}

func (nativeProbe) ID() string               { return "ArrayFlipInvalidValueType" }
func (nativeProbe) Semantic()                {}
func (nativeProbe) Flow()                    {}
func (nativeProbe) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (p nativeProbe) Check(ctx *analysis.Context, n syntax.Node) { p.check(ctx, n.(*syntax.FuncCall)) }

func probeNative(t *testing.T, src string, check func(*analysis.Context, *syntax.FuncCall)) {
	t.Helper()
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		if name, ok := c.Name.(*syntax.Name); ok && name.Value == "probe" {
			check(ctx, c)
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	e.Analyze(syntax.Parse("test.php", []byte("<?php "+src), syntax.Options{}))
}

func TestNativeArgumentBinding(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"probe(4, 8);", "4"},
		{"probe(value: 7);", "7"},
		{"probe(...$args);", ""},
		{"probe(4, value: 7);", ""},
		{"probe(other: 7);", ""},
	} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				if got := ctx.Text(CallArgument(c.Args, 0, "value")); got != tc.want {
					t.Fatalf("got %q, want %q", got, tc.want)
				}
			})
		})
	}
	if CallArgument(nil, 0, "value") != nil {
		t.Fatal("nil arguments must stay unknown")
	}
}

func TestNativeScalarContracts(t *testing.T) {
	for _, tc := range []struct {
		src          string
		truth, known bool
	}{
		{"probe(true);", true, true},
		{"probe(false);", false, true},
		{"probe(null);", false, true},
		{"probe(0);", false, true},
		{"probe(-4);", true, true},
		{"probe('0');", false, true},
		{"probe('');", false, true},
		{"probe('yes');", true, true},
		{"probe(0.0);", false, true},
		{"probe(1.25);", true, true},
		{"probe($unknown);", false, false},
		{"probe([]);", false, false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				got, known := NativeTruth(ctx, CallArgument(c.Args, 0, "value"))
				if got != tc.truth || known != tc.known {
					t.Fatalf("got (%v,%v), want (%v,%v)", got, known, tc.truth, tc.known)
				}
			})
		})
	}
	probeNative(t, "probe(1 | 4);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if present, known := NativeFlag(ctx, CallArgument(c.Args, 0, "value"), 4); !present || !known {
			t.Fatal("known bit-or flags must resolve")
		}
		if present, known := NativeFlag(ctx, nil, 4); present || !known {
			t.Fatal("omitted flags default to zero")
		}
	})
	probeNative(t, "probe(1 | $unknown);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		_, known := NativeFlag(ctx, CallArgument(c.Args, 0, "value"), 4)
		if known {
			t.Fatal("dynamic flags must remain unknown")
		}
	})
}

func TestNativeArrayContracts(t *testing.T) {
	for _, tc := range []struct {
		src   string
		count int
		known bool
	}{
		{"probe(['a'=>1,'a'=>2]);", 1, true},
		{"probe([4=>1,2]);", 2, true},
		{"probe(['4'=>1,4=>2]);", 1, true},
		{"probe(['04'=>1,4=>2]);", 2, true},
		{"probe([$key=>1]);", 0, false},
		{"probe([...$items]);", 0, false},
		{"probe([&$v]);", 0, false},
		{"probe(4);", 0, false},
		{"probe([PHP_INT_MAX=>1,2]);", 0, false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				entries, known := NativeArrayEntries(ctx, CallArgument(c.Args, 0, "value"))
				if known != tc.known || (known && len(entries) != tc.count) {
					t.Fatalf("got count %d known %v; want %d,%v", len(entries), known, tc.count, tc.known)
				}
			})
		})
	}
}

func TestNativeCallbackContracts(t *testing.T) {
	for _, tc := range []struct {
		src      string
		count    int
		complete bool
	}{
		{"probe(fn($a)=>$a);", 1, true},
		{"probe(function($a){if($a){return 1;}});", 0, false},
		{"probe(function($a){return $a;});", 1, true},
		{"probe(function(){return;});", 0, false},
		{"probe(function(){yield 1;});", 0, false},
		{"probe(function(){ $f=function(){return 1;}; });", 0, false},
		{"probe($unknown);", 0, false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				_, body := NativeCallback(ctx, CallArgument(c.Args, 0, "value"))
				values, complete := NativeReturns(body)
				if len(values) != tc.count || complete != tc.complete {
					t.Fatalf("got returns %d complete %v, want %d %v", len(values), complete, tc.count, tc.complete)
				}
			})
		})
	}
	if NativeGeneratorBody(nil) {
		t.Fatal("nil body isn't a generator")
	}
	probeNative(t, "function values(){yield 2;} probe(values());", func(ctx *analysis.Context, c *syntax.FuncCall) {
		inner := CallArgument(c.Args, 0, "value").(*syntax.FuncCall)
		body := NativeFunctionBody(ctx, inner)
		if !NativeGeneratorBody(body) {
			t.Fatal("same-file generator must resolve")
		}
	})
}

func TestNativeFailurePolicyGuards(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{"probe();", true},
		{"probe(0);", true},
		{"probe(FILTER_NULL_ON_FAILURE);", false},
		{"probe(['flags'=>FILTER_NULL_ON_FAILURE]);", false},
		{"probe(['options'=>['min_range'=>1]]);", false},
		{"probe(['options'=>['max_range'=>-1]]);", false},
		{"probe(['options'=>['min_range'=>0,'max_range'=>5]]);", true},
		{"probe(['options'=>$unknown]);", false},
		{"probe(['flags'=>$unknown]);", false},
		{"probe(['options'=>['min_range'=>$unknown]]);", false},
		{"probe($unknown);", false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				got := NativeIntegerAllowsZero(ctx, CallArgument(c.Args, 0, "value"))
				if got != tc.want {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			})
		})
	}
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{"if(strlen($s)===4){probe($s);}", true},
		{"if(strlen($s)===8){probe($s);}", false},
		{"if(strlen($other)===4){probe($s);}", false},
		{"if(count($s)===4){probe($s);}", false},
		{"if(strlen($s)!==4){return;} probe($s);", true},
		{"if(strlen($s)!==4){echo 'short';} probe($s);", false},
		{"probe($s);", false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				got := NativeLengthGuard(ctx, CallArgument(c.Args, 0, "value"), 4)
				if got != tc.want {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			})
		})
	}
}

func TestNativeCurlOptionProof(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{"$h=curl_init('https://example.invalid'); probe($h);", true},
		{"$h=curl_init('https://example.invalid'); curl_setopt($h,CURLOPT_RETURNTRANSFER,false); probe($h);", true},
		{"$h=curl_init('https://example.invalid'); curl_setopt($h,CURLOPT_RETURNTRANSFER,true); probe($h);", false},
		{"$h=curl_init('https://example.invalid'); curl_setopt($h,CURLOPT_RETURNTRANSFER,$unknown); probe($h);", false},
		{"$h=curl_init('https://example.invalid'); curl_setopt($h,$option,false); probe($h);", false},
		{"$h=curl_init('https://example.invalid'); curl_setopt($h,CURLOPT_WRITEFUNCTION,$callback); probe($h);", false},
		{"$h=curl_init('https://example.invalid'); curl_setopt_array($h,[]); probe($h);", false},
		{"$h=curl_init('https://example.invalid'); retain($h); probe($h);", false},
		{"probe($unknown);", false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				got := NativeCurlDefault(ctx, c, CallArgument(c.Args, 0, "value"), "CURLOPT_RETURNTRANSFER")
				if got != tc.want {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			})
		})
	}
}

func TestNativeStringInputContracts(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"probe(trim($s));", 1},
		{"probe(trim(string:$s));", 1},
		{"probe(trim(...$args));", 0},
		{"probe($unknown($s));", 0},
		{"function custom(?string $s){} probe(custom($s));", 0},
		{"function custom(string $s){} probe(custom($s));", 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				inner := CallArgument(c.Args, 0, "value").(*syntax.FuncCall)
				got := NativeStringInputs(ctx, inner, inner.Args)
				if len(got) != tc.want {
					t.Fatalf("got %d, want %d", len(got), tc.want)
				}
			})
		})
	}
}

func TestNativeSentinelGuards(t *testing.T) {
	for _, tc := range []struct {
		src, sentinel string
		want          bool
	}{
		{"if($v!==null){probe($v);}", "null", true},
		{"if(null!==$v){probe($v);}", "null", true},
		{"if($v===null){return;} probe($v);", "null", true},
		{"if($v===false){}else{probe($v);}", "false", true},
		{"if($v!==false){probe($v);}", "null", false},
		{"if($other!==null){probe($v);}", "null", false},
		{"if($v){probe($v);}", "null", false},
		{"if($v==null){return;} probe($v);", "null", false},
		{"probe();", "null", false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				got := NativeSentinelGuard(ctx, CallArgument(c.Args, 0, "value"), tc.sentinel)
				if got != tc.want {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			})
		})
	}
}

func TestNativeSuccessPromises(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{"function save(){probe();return true;}", true},
		{"function save(){probe();return false;}", false},
		{"function save(){probe();return $unknown;}", false},
		{"function save(){probe();echo 'done';}", false},
		{"probe();", false},
		{"$r=probe();", false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				got := NativeSuccessFollower(ctx, c)
				if got != tc.want {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			})
		})
	}
}

func TestNativeNamedCallbacks(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{"function cb($v){return $v;} probe('cb');", true},
		{"probe('missing');", false},
		{"probe('strtolower');", false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
				_, body := NativeCallback(ctx, CallArgument(c.Args, 0, "value"))
				if (body != nil) != tc.want {
					t.Fatalf("body exists %v, want %v", body != nil, tc.want)
				}
			})
		})
	}
	probeNative(t, "probe(unresolved());", func(ctx *analysis.Context, c *syntax.FuncCall) {
		inner := CallArgument(c.Args, 0, "value").(*syntax.FuncCall)
		if NativeFunctionBody(ctx, inner) != nil {
			t.Fatal("unresolved body must remain absent")
		}
	})
	probeNative(t, "function outer(){ $f=function(){yield 1;}; return 1; } probe(outer());", func(ctx *analysis.Context, c *syntax.FuncCall) {
		inner := CallArgument(c.Args, 0, "value").(*syntax.FuncCall)
		if NativeGeneratorBody(NativeFunctionBody(ctx, inner)) {
			t.Fatal("nested generator must not mark its enclosing function")
		}
	})
}

func TestNativeRemainingBoundaries(t *testing.T) {
	probeNative(t, "$v=7;probe($v);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if n, known := NativeInt(ctx, CallArgument(c.Args, 0, "value")); !known || n != 7 {
			t.Fatal("dominated scalar assignment must resolve")
		}
	})
	probeNative(t, "probe($flag?1:2);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if NativeValue(ctx, CallArgument(c.Args, 0, "value")) != nil {
			t.Fatal("alternative literal values must remain distinct")
		}
	})
	probeNative(t, "probe(~4);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if _, known := NativeInt(ctx, CallArgument(c.Args, 0, "value")); known {
			t.Fatal("non-sign unary expressions are not integer literals")
		}
	})
	probeNative(t, "if($v){probe($v);}", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if NativeLengthGuard(ctx, CallArgument(c.Args, 0, "value"), 4) {
			t.Fatal("truthiness does not prove a byte length")
		}
	})
	probeNative(t, "function f($v){probe($v);}", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if NativeSentinelGuard(ctx, CallArgument(c.Args, 0, "value"), "null") {
			t.Fatal("parameter presence does not prove nonnull")
		}
	})
	probeNative(t, "$h=curl_init('https://example.invalid'); $o=new stdClass(); $o->retain($h);probe($h);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if NativeCurlDefault(ctx, c, CallArgument(c.Args, 0, "value"), "CURLOPT_RETURNTRANSFER") {
			t.Fatal("unknown receiver escape must not prove default")
		}
	})
	probeNative(t, "probe(function(){ $f=function(){return 1;};return 2;});", func(ctx *analysis.Context, c *syntax.FuncCall) {
		_, b := NativeCallback(ctx, CallArgument(c.Args, 0, "value"))
		returns, complete := NativeReturns(b)
		if !complete || len(returns) != 1 {
			t.Fatal("nested callback returns must stay separate")
		}
	})
	probeNative(t, "probe(function(){yield 1;return 2;});", func(ctx *analysis.Context, c *syntax.FuncCall) {
		_, b := NativeCallback(ctx, CallArgument(c.Args, 0, "value"))
		_, complete := NativeReturns(b)
		if complete {
			t.Fatal("generator callback does not return mapped values")
		}
	})
}
