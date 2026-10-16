package infer

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

func assertionEnv(t *testing.T, src string, enabled bool, php phpversion.Version) (*Env, *syntax.File) {
	t.Helper()
	f := syntax.Parse("test.php", []byte("<?php "+src), syntax.Options{Version: phpversion.Max})
	if len(f.Errors) > 0 {
		t.Fatal(f.Errors)
	}
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	return NewEnv(f, names.New(f), ix, php, Options{Assertions: enabled}), f
}

func assertionRead(f *syntax.File) syntax.Expr {
	var read syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if call, ok := n.(*syntax.FuncCall); ok {
			if name, ok := call.Name.(*syntax.Name); ok && name.Value == "probe" {
				read = call.Args.Args[0].(*syntax.Arg).Value
			}
		}
		return true
	})
	return read
}

func TestAssertionInferenceOption(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		e, f := assertionEnv(t, `function f(?string $v){assert(is_string($v));probe($v);}`, enabled, phpversion.PHP84)
		got := e.TypeOf(assertionRead(f)).String()
		want := "null|string"
		if enabled {
			want = "string"
		}
		if got != want {
			t.Fatalf("enabled=%v type=%s", enabled, got)
		}
		if e.Native().TypeOf(assertionRead(f)).String() != want {
			t.Fatal("native twin lost assertion mode")
		}
		opposite := e.WithOptions(Options{Assertions: !enabled})
		if opposite == e || opposite.options.Assertions == enabled {
			t.Fatal("changed options reused cached environment")
		}
		if e.WithOptions(e.options) != e {
			t.Fatal("unchanged options should retain cache")
		}
	}
}

func TestOptionalAssertionDefinitions(t *testing.T) {
	for _, src := range []string{
		`$v=$_GET['x'];assert(($v=(int)$v)>=0);probe($v);`,
		`$v=1;assert(($v=$_GET['x'])!=='');probe($v);`,
		`$v=1;do{probe($v);assert(($v=$_GET['x'])!=='');}while($c);`,
		`$v=$_GET['x'];do{probe($v);assert(($v=(int)$v)>=0);}while($c);`,
		`$v='a';do{probe($v);assert(++$v);}while($c);`,
		`$v=$_GET['x'];assert(settype($v,'integer'));probe($v);`,
		`function rewrite(int &$x){} $v=$_GET['x'];assert(rewrite($v));probe($v);`,
		`class C{function rewrite(int &$x){}} $v=$_GET['x'];$o=new C;assert($o->rewrite($v));probe($v);`,
		`$v=['x'=>$_GET['x']];assert(($v['x']=(int)$v['x'])>=0);probe($v['x']);`,
	} {
		e, f := assertionEnv(t, `function f(){`+src+`}`, false, phpversion.PHP84)
		got := e.TypeOf(assertionRead(f))
		if got.OnlyOf("int", "float", "bool", "true", "false", "null") {
			t.Fatalf("optional mutation supplied scalar proof: %s => %s", src, got.String())
		}
	}
	// PHP5 ordinary-function argument expressions run even without assumptions.
	e, f := assertionEnv(t, `$v='a';assert(($v=1)>0);probe($v);`, false, phpversion.PHP56)
	if e.TypeOf(assertionRead(f)).String() != "int|string" {
		t.Fatalf("PHP5 argument type=%s", e.TypeOf(assertionRead(f)).String())
	}
}

func TestLegacyAssertionQueryOrder(t *testing.T) {
	for _, src := range []string{`function f(int $n){assert($n);probe($n);}`, `function f(){ $n=1;assert($n);probe($n);}`} {
		for _, first := range []string{"probe", "argument", "native"} {
			e, f := assertionEnv(t, src, false, phpversion.PHP74)
			var arg syntax.Expr
			syntax.InspectFile(f, func(n syntax.Node) bool {
				if call, ok := n.(*syntax.FuncCall); ok && e.BuiltinAssert(call) {
					arg = call.Args.Args[0].(*syntax.Arg).Value
				}
				return true
			})
			switch first {
			case "argument":
				e.TypeOf(arg)
			case "native":
				e.Native().TypeOf(arg)
			}
			got := e.TypeOf(assertionRead(f))
			if !got.IsUnknown() && got.String() != "mixed" {
				t.Fatalf("%s: legacy variable assertion retained proof %s", first, got.String())
			}
		}
	}
}

func TestAssertionDescriptionCannotRewriteGuard(t *testing.T) {
	for _, description := range []string{`$v=null`, `rewrite($v)`, `$o->description`, `$missing`, `UNDEFINED_DESCRIPTION`} {
		e, f := assertionEnv(t, `function f(?string $v){assert(is_string($v), `+description+`);probe($v);}`, true, phpversion.PHP84)
		if e.TypeOf(assertionRead(f)).OnlyOf("string") {
			t.Fatalf("description %s retained stale guard", description)
		}
	}
}

func TestAssertionArgumentBinding(t *testing.T) {
	for _, tc := range []struct {
		call  string
		guard bool
	}{
		{`assert(true,"message")`, true},
		{`assert(true,null)`, true},
		{`assert(assertion:true,description:"message")`, true},
		{`assert(description:rewrite(),assertion:true)`, true},
		{`assert(assertion:true,true)`, false},
		{`assert(assertion:true,assertion:false)`, false},
		{`assert(...)`, false},
		{`assert()`, false},
		{`assert(...$args)`, false},
	} {
		e, f := assertionEnv(t, tc.call+`;`, true, phpversion.PHP84)
		var call *syntax.FuncCall
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok {
				if name, ok := c.Name.(*syntax.Name); ok && name.Value == "assert" {
					call = c
				}
			}
			return true
		})
		if got := e.AssertGuard(call) != nil; got != tc.guard {
			t.Fatalf("enabled guard %s=%v", tc.call, got)
		}
		if e.WithOptions(Options{}).AssertGuard(call) != nil {
			t.Fatal("disabled guard")
		}
		if got := e.AssertCondition(call) != nil; got != tc.guard {
			t.Fatalf("%s guard=%v", tc.call, got)
		}
	}
}

func TestLegacyAssertionResultKinds(t *testing.T) {
	for _, tc := range []struct {
		condition string
		code      bool
	}{
		{`1`, false},
		{`1.5`, false},
		{`'code'`, true},
		{`true`, false},
		{`null`, false},
		{`UNKNOWN`, true},
		{`!$v`, false},
		{`(int)$v`, false},
		{`-$v`, true},
		{`$v+1`, true},
		{`$v!==false`, false},
		{`[]`, false},
		{`new stdClass`, false},
		{`isset($v)`, false},
		{`empty($v)`, false},
		{`$v instanceof stdClass`, false},
		{`is_string($v)`, false},
		{`strlen($v)`, false},
		{`htmlspecialchars($v)`, true},
		{`unknown()`, true},
		{``, true},
	} {
		e, f := assertionEnv(t, `assert(`+tc.condition+`);`, false, phpversion.PHP74)
		var call *syntax.FuncCall
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok && e.BuiltinAssert(c) {
				call = c
			}
			return true
		})
		if got := e.LegacyAssertCode(call); got != tc.code {
			t.Fatalf("%s legacycode=%v", tc.condition, got)
		}
	}
}

func TestOptionalAssertionInvalidatesEarlierGuard(t *testing.T) {
	for _, body := range []string{
		`if(!is_int($v)){return;}assert(($v=$_GET['x'])!=='');probe($v);`,
		`if(!is_int($v)){return;}assert(settype($v,'string'));probe($v);`,
		`function rewrite(string &$x){} if(!is_int($v)){return;}assert(rewrite($v));probe($v);`,
		`if(!is_int($v['x'])){return;}assert(($v['x']=$_GET['x'])!=='');probe($v['x']);`,
	} {
		e, f := assertionEnv(t, `function f($v){`+body+`}`, false, phpversion.PHP84)
		if got := e.TypeOf(assertionRead(f)); got.OnlyOf("int") {
			t.Fatalf("stale early guard: %s", body)
		}
	}
}

func TestAssertionConditionRewritesTestedValue(t *testing.T) {
	e, f := assertionEnv(t, `function f(int $x){assert(is_int($x)&&($x=$_GET['x']));probe($x);}`, true, phpversion.PHP84)
	if got := e.TypeOf(assertionRead(f)); got.OnlyOf("int") {
		t.Fatal("condition retained fact from before its assignment")
	}
}

func TestAssertionGuardRejectsInternalEffects(t *testing.T) {
	for _, condition := range []string{`is_int($x)&&($x=$_GET['x'])`, `($x=1)&&is_int($x)`, `is_int($x)&&unknown()`, `is_int($x)&&strlen($x)`, `$o->check($x)`, `++$x`, `is_int($x)&&$o->flag`, `is_int($x)&&eval('$x=5;return true;')`, `is_int($x)&&include('x.php')`, `is_int($x)&&is_bool($o->flag)`, `is_int($x)&&is_bool($o[0])`, `is_int(...$args)`, `is_int(&$x)`, `$x>0`, `+ $x`, `${$name}`, `UNDEFINED`, `is_int($x)&&$o->flag&&is_bool($x)`} {
		e, f := assertionEnv(t, `function f($x){assert(`+condition+`);}`, true, phpversion.PHP84)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok && e.BuiltinAssert(c) {
				if e.AssertGuard(c) != nil {
					t.Fatalf("effectful guard: %s", condition)
				}
			}
			return true
		})
	}
}

func TestAssertionGuardSafeReads(t *testing.T) {
	for _, condition := range []string{`is_int(value:$x)`, `!empty($x)`, `isset($x)`, `$x===null`, `$x!==false`, `is_iterable($x)`, `$x instanceof stdClass`, `($x)`} {
		e, f := assertionEnv(t, `function f($x){assert(`+condition+`);}`, true, phpversion.PHP84)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok && e.BuiltinAssert(c) {
				if e.AssertGuard(c) == nil {
					t.Fatalf("safe guard rejected: %s", condition)
				}
			}
			return true
		})
	}
}

func TestLegacyAssertionCannotSupplySharedGuard(t *testing.T) {
	for _, tc := range []struct {
		condition string
		guard     bool
	}{{`$x`, false}, {`is_int($x)`, true}, {`$x!==false`, true}, {`'$x=false;return true;'`, false}} {
		e, f := assertionEnv(t, `function f($x){assert(`+tc.condition+`);}`, true, phpversion.PHP74)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok && e.BuiltinAssert(c) {
				if got := e.AssertGuard(c) != nil; got != tc.guard {
					t.Fatalf("%s guard=%v", tc.condition, got)
				}
			}
			return true
		})
	}
}

func TestOptionalAssertionsInvalidateTypedAndElementProofs(t *testing.T) {
	for _, src := range []string{
		`function f(int $v){if(!is_int($v)){return;}assert(($v=$_GET['x'])!=='');probe($v);}`,
		`function f(){ $v=['x'=>1];assert(++$v['x']);probe($v['x']);}`,
		`function f(){ $v='raw';if(assert(($v=1)>0)){probe($v);}}`,
	} {
		e, f := assertionEnv(t, src, false, phpversion.PHP84)
		if got := e.TypeOf(assertionRead(f)); got.OnlyOf("int") {
			t.Fatalf("optional write established integer proof: %s", src)
		}
	}
}
