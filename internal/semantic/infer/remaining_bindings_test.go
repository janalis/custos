package infer

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func TestTemplateBindingEvidence(t *testing.T) {
	const prefix = `class C{} /**
 * @template T
 * @param T $value
 * @return T
 */ function identity($value, $option=null){} `
	for _, tc := range []struct {
		args  string
		php   phpversion.Version
		known bool
	}{
		{`new C`, phpversion.PHP84, true},
		{`new C,null,17`, phpversion.PHP84, true},
		{`value: new C`, phpversion.PHP84, true},
		{`option:null,value:new C`, phpversion.PHP84, true},
		{`VALUE:new C`, phpversion.PHP84, false},
		{`value:new C`, phpversion.PHP74, false},
		{`new C,value:new C`, phpversion.PHP84, false},
		{`value:new C,value:new C`, phpversion.PHP84, false},
		{`new C,...$tail`, phpversion.PHP84, true},
		{`...$tail,value:new C`, phpversion.PHP84, true},
		{`...$tail,value:new C`, phpversion.PHP80, false},
		{`...$tail,value:new C`, phpversion.PHP81, true},
		{`...$tail`, phpversion.PHP84, false},
		{`option:null,new C`, phpversion.PHP84, false},
	} {
		t.Run(tc.args+tc.php.String(), func(t *testing.T) {
			env, file := assertionEnv(t, prefix+`probe(identity(`+tc.args+`));`, false, tc.php)
			call := assertionRead(file).(*syntax.FuncCall)
			fn := env.ResolveFunction(call)
			if fn == nil || fn.Tpl == nil {
				t.Fatal("missing template metadata")
			}
			_, known := env.tplReturn(fn.Tpl, fn.Params, call.Args, fn.Return, nil, nil)
			if known != tc.known {
				t.Fatalf("known=%v want=%v", known, tc.known)
			}
		})
	}
}

func TestConditionalUnknownBindingIsNotDefault(t *testing.T) {
	const prefix = `/** @return ($flag is true ? int : string) */ function choose(bool $flag=false){} `
	for _, args := range []string{`FLAG:true`, `true,flag:false`, `...$args`, `flag:true,flag:false`} {
		env, file := assertionEnv(t, prefix+`probe(choose(`+args+`));`, false, phpversion.PHP84)
		value := assertionRead(file)
		if got := env.TypeOf(value).String(); got != "int|string" {
			t.Errorf("%s: got %s", args, got)
		}
	}
}

func TestTRulesBindingCase(t *testing.T) {
	for _, tc := range []struct {
		args  string
		bound bool
	}{
		{`'x',return:true`, true}, {`'x',RETURN:true`, false}, {`'x',true,return:false`, false},
	} {
		env, file := assertionEnv(t, `probe(print_r(`+tc.args+`));`, false, phpversion.PHP84)
		call := assertionRead(file).(*syntax.FuncCall)
		fn := env.ResolveFunction(call)
		if got := NewTRules(env).arg(call, fn, 1) != nil; got != tc.bound {
			t.Errorf("%s: bound=%v", tc.args, got)
		}
	}
}

func TestParameterForExactName(t *testing.T) {
	params := []index.Param{{Name: "$value", ByRef: true}}
	if _, known := paramFor(params, &syntax.Arg{Name: &syntax.Identifier{Value: "VALUE"}}, 0); known {
		t.Fatal("case mismatch bound by reference")
	}
}

func TestArgumentBindingStates(t *testing.T) {
	value := &syntax.Variable{Name: "value"}
	arg := func(name string, unpack bool) *syntax.Arg {
		a := &syntax.Arg{Value: value, Unpack: unpack}
		if name != "" {
			a.Name = &syntax.Identifier{Value: name}
		}
		return a
	}
	params := []index.Param{{Name: "$value"}, {Name: "$option"}}
	for _, tc := range []struct {
		name   string
		args   *syntax.ArgList
		params []index.Param
		target int
		php    phpversion.Version
		want   argumentBindingState
	}{
		{"nil", nil, params, 0, 0, bindingUnknown},
		{"negative", &syntax.ArgList{}, params, -1, 0, bindingUnknown},
		{"outside", &syntax.ArgList{}, params, 2, 0, bindingUnknown},
		{"parameter budget", &syntax.ArgList{}, make([]index.Param, 65), 0, 0, bindingUnknown},
		{"argument budget", &syntax.ArgList{Args: make([]syntax.Expr, 257)}, params, 0, 0, bindingUnknown},
		{"variadic target", &syntax.ArgList{}, []index.Param{{Variadic: true}}, 0, 0, bindingUnknown},
		{"missing", &syntax.ArgList{}, params, 0, 0, bindingMissing},
		{"placeholder", &syntax.ArgList{Args: []syntax.Expr{value}}, params, 0, 0, bindingUnknown},
		{"nil argument", &syntax.ArgList{Args: []syntax.Expr{(*syntax.Arg)(nil)}}, params, 0, 0, bindingUnknown},
		{"nil value", &syntax.ArgList{Args: []syntax.Expr{&syntax.Arg{}}}, params, 0, 0, bindingUnknown},
		{"old unpack", &syntax.ArgList{Args: []syntax.Expr{arg("", true)}}, params, 0, phpversion.PHP53, bindingUnknown},
		{"unpack after named", &syntax.ArgList{Args: []syntax.Expr{arg("value", false), arg("", true)}}, params, 0, 0, bindingUnknown},
		{"positional after unpack", &syntax.ArgList{Args: []syntax.Expr{arg("", true), arg("", false)}}, params, 0, 0, bindingUnknown},
		{"extra positional", &syntax.ArgList{Args: []syntax.Expr{arg("", false), arg("", false), arg("", false)}}, params, 0, 0, bindingBound},
		{"variadic overflow", &syntax.ArgList{Args: []syntax.Expr{arg("", false), arg("", false), arg("", false)}}, []index.Param{{Name: "value"}, {Name: "tail", Variadic: true}}, 0, 0, bindingBound},
		{"variadic named", &syntax.ArgList{Args: []syntax.Expr{arg("other", false), arg("value", false)}}, []index.Param{{Name: "value"}, {Name: "tail", Variadic: true}}, 0, 0, bindingBound},
		{"unpack missing", &syntax.ArgList{Args: []syntax.Expr{arg("", true)}}, params, 0, 0, bindingUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, state := bindArgument(tc.args, tc.params, tc.target, tc.php)
			if state != tc.want || (got != nil) != (state == bindingBound) {
				t.Fatalf("state=%v value=%v want=%v", state, got, tc.want)
			}
		})
	}
}

func BenchmarkArgumentBinding(b *testing.B) {
	params := []index.Param{{Name: "value"}, {Name: "option"}}
	args := &syntax.ArgList{Args: []syntax.Expr{&syntax.Arg{Value: &syntax.Variable{Name: "v"}}, &syntax.Arg{Value: &syntax.Variable{Name: "tail"}, Unpack: true}}}
	b.ReportAllocs()
	for b.Loop() {
		bindArgument(args, params, 0, phpversion.PHP84)
	}
}

func TestConditionalMissingBinding(t *testing.T) {
	for _, tc := range []struct{ parameter, want string }{{"bool $flag=false", "string"}, {"bool $flag", "int|string"}} {
		env, file := assertionEnv(t, `/** @return ($flag is true ? int : string) */ function choose(`+tc.parameter+`){} probe(choose());`, false, phpversion.PHP84)
		if got := env.TypeOf(assertionRead(file)).String(); got != tc.want {
			t.Fatalf("%s: %s want %s", tc.parameter, got, tc.want)
		}
	}
	if _, known := paramFor([]index.Param{{Variadic: true}}, &syntax.Arg{}, -1); known {
		t.Fatal("negative position bound variadic")
	}
}
