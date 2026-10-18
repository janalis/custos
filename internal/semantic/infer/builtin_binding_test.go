package infer

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func TestBuiltinBindingProof(t *testing.T) {
	for _, tc := range []struct {
		src  string
		php  phpversion.Version
		want bool
	}{
		{`probe(gettimeofday(as_float: true));`, phpversion.PHP84, true},
		{`probe(gettimeofday(AS_FLOAT: true));`, phpversion.PHP84, false},
		{`probe(gettimeofday(as_float: true));`, phpversion.PHP74, false},
		{`probe(gettimeofday(...$args, as_float: true));`, phpversion.PHP84, false},
		{`probe(gettimeofday(true, as_float: false));`, phpversion.PHP84, false},
		{`probe(gettimeofday(true, false));`, phpversion.PHP84, false},
		{`probe(gettimeofday(as_float: true, false));`, phpversion.PHP84, false},
		{`probe(gettimeofday(...));`, phpversion.PHP84, false},
		{`function gettimeofday($v){return 'local';}probe(gettimeofday(true));`, phpversion.PHP84, true},
		{`probe(str_replace(subject: [], search: 'a', replace: 'b'));`, phpversion.PHP84, true},
		{`probe(max(value: [1]));`, phpversion.PHP84, false},
		{`probe(str_replace(subject: []));`, phpversion.PHP84, false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, f := assertionEnv(t, tc.src, false, tc.php)
			call := assertionRead(f).(*syntax.FuncCall)
			_, got := e.overrideType(call, call.Name.(*syntax.Name))
			if got != tc.want {
				t.Fatalf("override=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestBuiltinBindingsKnown(t *testing.T) {
	params := []index.Param{{Name: "$value"}, {Name: "option", Optional: true}}
	for _, tc := range []struct {
		args       string
		params     []index.Param
		php        phpversion.Version
		tail, want bool
	}{
		{"1", params, phpversion.PHP84, false, true},
		{"option: 2, value: 1", params, phpversion.PHP84, false, true},
		{"value: 1, value: 2", params, phpversion.PHP84, false, false},
		{"Value: 1", params, phpversion.PHP84, false, false},
		{"&$x", params, phpversion.PHP53, false, false},
		{"1, 2, 3", nil, phpversion.PHP84, false, false},
		{"1, 2, 3", []index.Param{{Name: "values", Variadic: true}}, phpversion.PHP84, false, true},
		{"1, values: 2", []index.Param{{Name: "values", Variadic: true}}, phpversion.PHP84, false, false},
		{"1, ...$tail", params, phpversion.PHP84, true, true},
		{"value: 1, ...$tail", params, phpversion.PHP84, true, false},
		{"1, ...$tail", params, phpversion.PHP55, true, false},
		{"1, ...$tail, ...$other", params, phpversion.PHP84, true, false},
		{"...$tail", params, phpversion.PHP84, true, false},
		{strings.Repeat("1,", 257), params, phpversion.PHP84, false, false},
		{"", make([]index.Param, 65), phpversion.PHP84, false, false},
	} {
		t.Run(tc.args, func(t *testing.T) {
			file := syntax.Parse("binding.php", []byte("<?php f("+tc.args+");"), syntax.Options{})
			list := file.Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.FuncCall).Args
			if got := builtinBindingsKnown(list, tc.params, tc.php, tc.tail); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
	if builtinBindingsKnown(nil, params, phpversion.PHP84, false) {
		t.Fatal("nil list established binding")
	}
	if builtinBindingsKnown(&syntax.ArgList{Args: []syntax.Expr{&syntax.Arg{}}}, params, phpversion.PHP84, false) {
		t.Fatal("missing value established binding")
	}
}

func TestBuiltinOverrideRequiresDeclaration(t *testing.T) {
	for _, source := range []string{`probe(gettimeofday(true));`, `function gettimeofday($x){return 'local';}probe(gettimeofday(true));`} {
		e, f := assertionEnv(t, source, false, phpversion.PHP84)
		e.Index = index.New(nil)
		e.Index.Add(index.Extract(f))
		call := assertionRead(f).(*syntax.FuncCall)
		if _, ok := e.overrideType(call, call.Name.(*syntax.Name)); ok {
			t.Fatal("unresolved or user function received builtin contract")
		}
	}
}

func BenchmarkBuiltinBinding(b *testing.B) {
	f := syntax.Parse("binding.php", []byte("<?php f(option: 2, value: 1);"), syntax.Options{})
	list := f.Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.FuncCall).Args
	params := []index.Param{{Name: "value"}, {Name: "option", Optional: true}}
	b.ReportAllocs()
	for b.Loop() {
		builtinBindingsKnown(list, params, phpversion.PHP84, false)
	}
}

// Public index data can omit optional parameter metadata. Keep omitted slots
// conservative without indexing past that declaration.
func TestBuiltinIncompleteParameterMetadata(t *testing.T) {
	e, f := assertionEnv(t, `probe(print_r(1));`, false, phpversion.PHP84)
	e.Index = index.New(nil)
	e.Index.Add(&index.FileSymbols{Path: "metadata.php", Functions: []*index.Function{{FQN: "print_r", Builtin: true, Params: []index.Param{{Name: "value"}}}}})
	call := assertionRead(f).(*syntax.FuncCall)
	if got, ok := e.overrideType(call, call.Name.(*syntax.Name)); !ok || got.String() != "true" {
		t.Fatalf("got %s overridden=%v", got.String(), ok)
	}
}
