package infer

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

func TestPowerLiteralProofs(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{"2 ** 3", true}, {"2 ** 63", false}, {"2 ** 64", false}, {"2 ** -1", false}, {"$n ** 2", false}, {"2 ** $n", false}, {"1 ** 999", true}, {"0 ** 999", true}, {"(-1) ** 999", true}, {"2 ** 0", true}, {"(-2) ** 63", true}, {"999999999999999999999999 ** 2", false}, {"--2 ** 2", false},
	} {
		f := syntax.Parse("p.php", []byte("<?php "+tc.src+";"), syntax.Options{})
		st, ok := f.Stmts[0].(*syntax.ExprStmt)
		if !ok {
			t.Fatal(tc.src)
		}
		power, ok := st.Expr.(*syntax.Binary)
		if !ok {
			continue
		}
		if got := powerFitsInt(power.Left, power.Right); got != tc.want {
			t.Errorf("%s = %v want %v", tc.src, got, tc.want)
		}
	}
}

func TestDeferredNativeOutput(t *testing.T) {
	f := syntax.Parse("native.php", []byte(`<?php
 class Service {/** @param-out string $value */ function fill(&$value) {}}
 function run() {$service=new Service;$value=1;$service->fill($value);echo $value;}
 `), syntax.Options{})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := NewEnv(f, names.New(f), ix, phpversion.PHP84).Native()
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if echo, ok := n.(*syntax.Echo); ok {
			if got := env.TypeOf(echo.Exprs[0]); !got.IsUnknown() {
				t.Fatal(got)
			}
		}
		return true
	})
}

func TestPre80ArrayMutationFailuresKeepInput(t *testing.T) {
	f := syntax.Parse("old.php", []byte(`<?php
 function run($unknown, array|string $union) {
  $number=1;sort($number);echo $number;
  $string='s';array_push($string,1);echo $string;
  sort($unknown);echo $unknown;
  sort($union);echo $union;
 }
 `), syntax.Options{})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := NewEnv(f, names.New(f), ix, phpversion.PHP74)
	want := []string{"int", "string", "?unknown", "?unknown"}
	i := 0
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if echo, ok := n.(*syntax.Echo); ok {
			if got := env.TypeOf(echo.Exprs[0]).String(); got != want[i] {
				t.Errorf("%d:%s want%s", i, got, want[i])
			}
			i++
		}
		return true
	})
	if i != len(want) {
		t.Fatal(i)
	}
}

func TestConstructorReferenceArgument(t *testing.T) {
	for _, byref := range []bool{false, true} {
		parameter := "$value"
		if byref {
			parameter = "&$value"
		}
		f := syntax.Parse("construct.php", []byte("<?php class Receiver {function __construct("+parameter+"){}} $value=['key'=>1]; new Receiver($value);"), syntax.Options{})
		ix := index.New(nil)
		ix.Add(index.Extract(f))
		env := NewEnv(f, names.New(f), ix, phpversion.PHP84)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if call, ok := n.(*syntax.New); ok {
				arg := call.Args.Args[0].(*syntax.Arg)
				if got := env.byRefArg(mutation{call: call, arg: arg, idx: 0}); got != byref {
					t.Errorf("constructor reference=%v want %v", got, byref)
				}
			}
			return true
		})
	}
}
