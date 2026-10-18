package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func TestReferenceArgumentContracts(t *testing.T) {
	for _, tc := range []struct {
		body   string
		params []index.Param
		want   bool
	}{
		{`f(X: $v);`, []index.Param{{Name: "$x", ByRef: true}}, false},
		{`f(extra: $v);`, []index.Param{{Name: "$rest", Variadic: true, ByRef: true}}, true},
		{`f(...$args);`, []index.Param{{Name: "$x", ByRef: true}}, true},
		{`f(...$args);`, []index.Param{{Name: "$x"}, {Name: "$later", ByRef: true}}, true},
		{`f(1, ...$args);`, []index.Param{{Name: "$x", ByRef: true}}, true},
		{`f(1, 2, $v);`, []index.Param{{Name: "$rest", Variadic: true, ByRef: true}}, true},
	} {
		t.Run(tc.body, func(t *testing.T) {
			call := FirstExpr(t, Parse(t, "<?php "+tc.body)).(*syntax.FuncCall)
			if got := ArgBindsByRef(call.Args, tc.params, true); got != tc.want {
				t.Fatalf("reference binding=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestVersionedConstructorContracts(t *testing.T) {
	for _, version := range []phpversion.Version{phpversion.PHP53, phpversion.PHP74, phpversion.PHP80} {
		for _, tc := range []struct {
			code   string
			known  bool
			legacy bool
		}{
			{`class Legacy{function Legacy(&$value){}} probe(new Legacy($x));`, true, true},
			{`class Legacy{function Legacy(&$value){}} class Child extends Legacy{} probe(new Child($x));`, true, true},
			{`class EmptyBase{} class Child extends EmptyBase{} probe(new Child($x));`, true, false},
			{`class Child extends Missing{} probe(new Child($x));`, false, false},
			{`class A extends B{} class B extends A{} probe(new A($x));`, false, false},
			{`trait Body{} class Child{use Body;} probe(new Child($x));`, false, false},
			{`if($flag){class Choice{}}else{class Choice{function __construct(&$value){}}}probe(new Choice($x));`, false, false},
			{`namespace App;class Legacy{function Legacy(&$value){}}probe(new Legacy($x));`, version != phpversion.PHP53, false},
		} {
			t.Run(version.String()+tc.code, func(t *testing.T) {
				p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
					callee, known := ResolveCallee(ctx, CallArgument(c.Args, 0, ""))
					if known != tc.known {
						t.Fatalf("known=%v, want %v", known, tc.known)
					}
					wantReference := tc.legacy && version < phpversion.PHP80
					if known && (len(callee.Params) > 0 && callee.Params[0].ByRef) != wantReference {
						t.Fatalf("unexpected constructor contract %+v", callee)
					}
				}}
				e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{PHP: version, Only: []string{p.ID()}})
				if err != nil {
					t.Fatal(err)
				}
				e.Analyze(syntax.Parse("legacy.php", []byte("<?php "+tc.code), syntax.Options{}))
			})
		}
	}
}

func TestCalleeDeclarationBudget(t *testing.T) {
	code := strings.Repeat(`if(!function_exists('target')){function target($value){}}`, 65) + `probe(target($x));`
	probeNative(t, code, func(ctx *analysis.Context, call *syntax.FuncCall) {
		if _, known := ResolveCallee(ctx, CallArgument(call.Args, 0, "")); known {
			t.Fatal("declaration budget exhaustion established a contract")
		}
	})
}

func TestArgumentBindingStates(t *testing.T) {
	for _, tc := range []struct {
		body  string
		state ArgumentState
		text  string
	}{
		{`f();`, ArgumentMissing, ""},
		{`f(other: 1);`, ArgumentMissing, ""},
		{`f(7);`, ArgumentBound, "7"},
		{`f(value: 8);`, ArgumentBound, "8"},
		{`f(Value: 8);`, ArgumentMissing, ""},
		{`f(...$args);`, ArgumentUnknown, ""},
		{`f(...);`, ArgumentUnknown, ""},
		{`f(1, value: 2);`, ArgumentUnknown, ""},
	} {
		t.Run(tc.body, func(t *testing.T) {
			f := Parse(t, "<?php "+tc.body)
			call := FirstExpr(t, f).(*syntax.FuncCall)
			arg, state := CallArgumentBinding(call.Args, 0, "$value")
			if state != tc.state || arg != nil && Text(f, arg) != tc.text {
				t.Fatalf("argument=%v state=%v, want text=%q state=%v", arg, state, tc.text, tc.state)
			}
			if omitted := CallArgumentOmitted(call.Args, 0, "$value"); omitted != (tc.state == ArgumentMissing) {
				t.Fatalf("omitted=%v for state %v", omitted, tc.state)
			}
		})
	}
}

func TestIndividualReferenceBinding(t *testing.T) {
	params := []index.Param{{Name: "$copy"}, {Name: "$target", ByRef: true}}
	for _, tc := range []struct {
		body         string
		index        int
		byRef, known bool
	}{
		{`f($a, $b);`, 0, false, true},
		{`f($a, $b);`, 1, true, true},
		{`f(&$a);`, 0, true, true},
		{`f(target: $a);`, 0, true, true},
		{`f(TARGET: $a);`, 0, false, false},
		{`f(...$a);`, 0, true, false},
		{`f(...$a, target: $b);`, 1, true, true},
		{`f(...$a, copy: $b);`, 1, false, true},
		{`f($a, $b, ...$c);`, 2, false, false},
	} {
		t.Run(tc.body, func(t *testing.T) {
			call := FirstExpr(t, Parse(t, "<?php "+tc.body)).(*syntax.FuncCall)
			byRef, known := ArgumentBindsByRef(call.Args, call.Args.Args[tc.index].(*syntax.Arg), params, true)
			if byRef != tc.byRef || known != tc.known {
				t.Fatalf("reference=%v known=%v, want %v %v", byRef, known, tc.byRef, tc.known)
			}
		})
	}
	if ref, known := ArgumentBindsByRef(nil, nil, params, true); ref || known {
		t.Fatal("missing argument has no binding")
	}
	list := &syntax.ArgList{Args: []syntax.Expr{&syntax.VariadicPlaceholder{}}}
	if ref, known := ArgumentBindsByRef(list, &syntax.Arg{}, params, true); ref || known {
		t.Fatal("an argument outside the list has no binding")
	}
}

func TestCalleeReceiverAgreement(t *testing.T) {
	for _, tc := range []struct {
		types, second string
		known         bool
	}{
		{"A|B", "&$value", true},
		{"A|B", "$value", false},
		{"A|B", "&$different", false},
		{"A|string", "&$value", false},
		{"A|Missing", "&$value", false},
	} {
		t.Run(tc.types+tc.second, func(t *testing.T) {
			probeNative(t, `class A{function run(&$value){}} class B{function run(`+tc.second+`){}}
function test(`+tc.types+` $object){probe($object->run($x));}`, func(ctx *analysis.Context, call *syntax.FuncCall) {
				_, known := ResolveCallee(ctx, CallArgument(call.Args, 0, ""))
				if known != tc.known {
					t.Fatalf("known=%v, want %v", known, tc.known)
				}
			})
		})
	}
}

func TestCalleeDeclarationAgreement(t *testing.T) {
	for _, tc := range []struct {
		second string
		known  bool
	}{
		{`&$value`, true},
		{`$value`, false},
		{`&$different`, false},
	} {
		probeNative(t, `if($flag){function target(&$value){}}else{function target(`+tc.second+`){}} probe(target($x));`, func(ctx *analysis.Context, call *syntax.FuncCall) {
			_, known := ResolveCallee(ctx, CallArgument(call.Args, 0, ""))
			if known != tc.known {
				t.Fatalf("known=%v want=%v", known, tc.known)
			}
		})
	}
}

func BenchmarkConstructorContract(b *testing.B) {
	p := nativeProbe{check: func(ctx *analysis.Context, call *syntax.FuncCall) {
		construction := CallArgument(call.Args, 0, "")
		b.ReportAllocs()
		for b.Loop() {
			ResolveCallee(ctx, construction)
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	e.Analyze(syntax.Parse("constructor.php", []byte(`<?php class Base{function __construct(&$value){}}class Child extends Base{}probe(new Child($x));`), syntax.Options{}))
}

func TestStringConsumerRejectsUnboundInput(t *testing.T) {
	probeNative(t, `function accepts(string $text){} probe(accepts(value: 1)); probe(missing());`, func(ctx *analysis.Context, call *syntax.FuncCall) {
		inner := CallArgument(call.Args, 0, "").(*syntax.FuncCall)
		if NativeStringConsumer(ctx, inner, inner.Args, call) || NativeStringConsumer(ctx, inner, nil, call) {
			t.Fatal("unbound input is not a string consumer")
		}
	})
}

func BenchmarkArgumentContracts(b *testing.B) {
	f := syntax.Parse("arguments.php", []byte(`<?php f($first, $second, extra: $third);`), syntax.Options{})
	call := f.Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.FuncCall)
	params := []index.Param{{Name: "$first"}, {Name: "$rest", ByRef: true, Variadic: true}}
	last := call.Args.Args[2].(*syntax.Arg)
	b.ReportAllocs()
	for b.Loop() {
		ArgBindsByRef(call.Args, params, true)
		ArgumentBindsByRef(call.Args, last, params, true)
		CallArgumentBinding(call.Args, 0, "$first")
	}
}

func TestResolvedConstructorContract(t *testing.T) {
	probeNative(t, `class ParentType {function __construct(&$value){}}
class ChildType extends ParentType {} probe(new ChildType($v));`, func(ctx *analysis.Context, call *syntax.FuncCall) {
		callee, known := ResolveCallee(ctx, CallArgument(call.Args, 0, ""))
		if !known || len(callee.Params) != 1 || !callee.Params[0].ByRef {
			t.Fatalf("constructor contract=%+v, known=%v", callee, known)
		}
	})
}

func TestNamedParameterArgument(t *testing.T) {
	call := FirstExpr(t, Parse(t, `<?php f(value: 7);`)).(*syntax.FuncCall)
	if arg := CallArgument(call.Args, 0, "$value"); arg == nil {
		t.Fatal("indexed parameter names must bind named arguments")
	}
}

func TestUnknownOptionalArgumentContracts(t *testing.T) {
	probeNative(t, `probe(header('X-Test: yes', replace: true, replace: false), header('X-Test: yes', ...$options), $db->execute(...$params));`, func(ctx *analysis.Context, probe *syntax.FuncCall) {
		header := CallArgument(probe.Args, 0, "").(*syntax.FuncCall)
		if _, _, _, known := ExpansionHeader(ctx, header); known {
			t.Fatal("unpacked replacement policy was treated as omitted")
		}
		if noContentHeaderKeepsStatus(ctx, CallArgument(probe.Args, 1, "").(*syntax.FuncCall)) {
			t.Fatal("unpacked response code preserved status")
		}
		execute := CallArgument(probe.Args, 2, "").(*syntax.MethodCall)
		if _, known := nativeBindings(ctx, execute); known {
			t.Fatal("unpacked bindings were treated as omitted")
		}
	})
}

func TestStringConsumerDuplicateDeclarations(t *testing.T) {
	probeNative(t, `if($flag){function consume(string $value){}}else{function consume(string $value){}}probe(consume($value));`, func(ctx *analysis.Context, probe *syntax.FuncCall) {
		call := CallArgument(probe.Args, 0, "").(*syntax.FuncCall)
		if NativeStringConsumer(ctx, call, call.Args, CallArgument(call.Args, 0, "")) {
			t.Fatal("duplicate declarations established a native string consumer")
		}
	})
}
