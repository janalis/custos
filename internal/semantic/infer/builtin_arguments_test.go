package infer

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

func TestBuiltinArguments(t *testing.T) {
	for _, tc := range []struct {
		args    string
		version phpversion.Version
		known   bool
	}{
		{`$h, CURLOPT_SSL_VERIFYPEER, false`, phpversion.PHP84, true},
		{`value: false, option: CURLOPT_SSL_VERIFYPEER, handle: $h`, phpversion.PHP84, true},
		{`...[$h, CURLOPT_SSL_VERIFYPEER, false]`, phpversion.PHP84, true},
		{`...['value' => false, 'handle' => $h, 'option' => CURLOPT_SSL_VERIFYPEER]`, phpversion.PHP84, true},
		{`...[$h], option: CURLOPT_SSL_VERIFYPEER, value: false`, phpversion.PHP81, true},
		{`...['handle' => $h], ...['option' => CURLOPT_SSL_VERIFYPEER, 'value' => false]`, phpversion.PHP84, true},
		{`...[$h, 1 => 99, 1 => CURLOPT_SSL_VERIFYPEER, false]`, phpversion.PHP84, true},
		{`...[-1 => $h, CURLOPT_SSL_VERIFYPEER, false]`, phpversion.PHP84, true},
		{`...[-1 => $h, CURLOPT_SSL_VERIFYPEER, false]`, phpversion.PHP82, true},
		{`...[$h, CURLOPT_SSL_VERIFYPEER, false]`, phpversion.PHP55, false},
		{`handle: $h, option: CURLOPT_SSL_VERIFYPEER, value: false`, phpversion.PHP74, false},
		{`...[$h], option: CURLOPT_SSL_VERIFYPEER, value: false`, phpversion.PHP80, false},
		{`...[$h], CURLOPT_SSL_VERIFYPEER, false`, phpversion.PHP84, false},
		{`...$args`, phpversion.PHP84, false},
		{`...[$h, ...$options]`, phpversion.PHP84, false},
		{`...[$h, $key => false]`, phpversion.PHP84, false},
		{`...['01' => $h, CURLOPT_SSL_VERIFYPEER, false]`, phpversion.PHP84, false},
		{`...['' => $h]`, phpversion.PHP84, false},
		{`...[9223372036854775807 => $h]`, phpversion.PHP84, false},
		{`handle: $h, ...['option' => 1]`, phpversion.PHP84, false},
		{`$h, CURLOPT_SSL_VERIFYPEER`, phpversion.PHP84, false},
		{`$h, CURLOPT_SSL_VERIFYPEER, false, 1`, phpversion.PHP84, false},
		{`handle: $h, handle: $h, option: CURLOPT_SSL_VERIFYPEER, value: false`, phpversion.PHP84, false},
		{`$h, option: CURLOPT_SSL_VERIFYPEER, false`, phpversion.PHP84, false},
		{`$h, unknown: 1, value: false`, phpversion.PHP84, false},
		{`...`, phpversion.PHP84, false},
	} {
		e, f := assertionEnv(t, `probe(curl_setopt(`+tc.args+`));`, false, tc.version)
		call := assertionRead(f).(*syntax.FuncCall)
		args, known := e.BuiltinArguments(call)
		if known != tc.known {
			t.Fatalf("%s (%s): %v", tc.args, tc.version, known)
		}
		if known && (len(args) != 3 || !strings.Contains(string(f.Src[args[1].Span().Start:args[1].Span().End]), "CURLOPT_SSL_VERIFYPEER")) {
			t.Fatalf("wrong option: %+v", args)
		}
	}
	for _, src := range []string{`namespace App; function curl_setopt($a,$b,$c) {} probe(curl_setopt(1,2,3));`, `probe(unknown(1));`, `probe(gettimeofday());`} {
		e, f := assertionEnv(t, src, false, phpversion.Max)
		args, known := e.BuiltinArguments(assertionRead(f).(*syntax.FuncCall))
		if strings.Contains(src, "gettimeofday") {
			if !known || len(args) != 1 || args[0] != nil {
				t.Fatal("optional argument")
			}
		} else if known {
			t.Fatalf("non-native target: %s (%+v)", src, e.ResolveFunction(assertionRead(f).(*syntax.FuncCall)))
		}
	}
}

func TestBuiltinArgumentBoundsAndRecovery(t *testing.T) {
	for _, written := range []string{
		strings.Repeat("1,", maxWrittenArgs+1),
		"...[" + strings.Repeat("1,", maxWrittenArgs+1) + "]",
		"...[$h, &$option, false]",
	} {
		e, f := assertionEnv(t, `probe(curl_setopt(`+written+`));`, false, phpversion.Max)
		if _, known := e.BuiltinArguments(assertionRead(f).(*syntax.FuncCall)); known {
			t.Fatal("invalid or oversized binding")
		}
	}
	e, f := assertionEnv(t, `probe(sprintf('%s','x'));`, false, phpversion.Max)
	if _, known := e.BuiltinArguments(assertionRead(f).(*syntax.FuncCall)); known {
		t.Fatal("variadic binding")
	}
	call := assertionRead(f).(*syntax.FuncCall)
	call.Args = nil
	if _, known := e.BuiltinArguments(call); known {
		t.Fatal("missing argument list")
	}
}

func BenchmarkBuiltinArguments(b *testing.B) {
	f := syntax.Parse("bench.php", []byte(`<?php curl_setopt(...['value'=>false,'handle'=>$h,'option'=>CURLOPT_SSL_VERIFYPEER]);`), syntax.Options{Version: phpversion.Max})
	ix := index.New(stubs.Index())
	e := NewEnv(f, names.New(f), ix, phpversion.Max, Options{})
	call := f.Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.FuncCall)
	b.ReportAllocs()
	for b.Loop() {
		e.BuiltinArguments(call)
	}
}

func TestBuiltinArgumentMissingValue(t *testing.T) {
	e, f := assertionEnv(t, `probe(curl_setopt($h,1,false));`, false, phpversion.Max)
	call := assertionRead(f).(*syntax.FuncCall)
	if _, known := e.BuiltinArguments(nil); known {
		t.Fatal("nil call")
	}
	call.Args.Args[0].(*syntax.Arg).Value = nil
	if _, known := e.BuiltinArguments(call); known {
		t.Fatal("recovered missing value bound")
	}
}

func TestNestedArgumentPacks(t *testing.T) {
	for _, tc := range []struct {
		args    string
		version phpversion.Version
		known   bool
	}{
		{`...[...[$h],...[CURLOPT_SSL_VERIFYPEER,false]]`, phpversion.PHP74, true},
		{`...[...[4=>$h],...[8=>CURLOPT_SSL_VERIFYPEER],false]`, phpversion.PHP84, true},
		{`...[...['handle'=>$h],...['option'=>0],...['option'=>CURLOPT_SSL_VERIFYPEER,'value'=>false]]`, phpversion.PHP81, true},
		{`...[...['handle'=>$h],'option'=>CURLOPT_SSL_VERIFYPEER,'value'=>false]`, phpversion.PHP80, false},
		{`...[...[$h],CURLOPT_SSL_VERIFYPEER,false]`, phpversion.PHP73, false},
		{`...[...$unknown,CURLOPT_SSL_VERIFYPEER,false]`, phpversion.PHP84, false},
		{`...[...[$h],...[9223372036854775807=>CURLOPT_SSL_VERIFYPEER],false]`, phpversion.PHP84, false},
		{`...[...[''=>$h],CURLOPT_SSL_VERIFYPEER,false]`, phpversion.PHP84, false},
		{`...[...['unknown'=>$h],CURLOPT_SSL_VERIFYPEER,false]`, phpversion.PHP84, false},
		{`...[9223372036854775806=>$h,...[CURLOPT_SSL_VERIFYPEER],false]`, phpversion.PHP84, false},
	} {
		e, f := assertionEnv(t, `probe(curl_setopt(`+tc.args+`));`, false, tc.version)
		args, known := e.BuiltinArguments(assertionRead(f).(*syntax.FuncCall))
		if known != tc.known {
			t.Fatalf("%s (%s): %v", tc.args, tc.version, known)
		}
		if known {
			option := args[1].(*syntax.ConstFetch)
			if option.Name.Value != "CURLOPT_SSL_VERIFYPEER" {
				t.Fatal("wrong nested option binding")
			}
		}
	}
	nested := `[$h,CURLOPT_SSL_VERIFYPEER,false]`
	for i := 0; i < 8; i++ {
		nested = `[...` + nested + `]`
	}
	for _, args := range []string{`...` + nested, `...[...[],` + strings.Repeat(`1,`, maxWrittenArgs) + `]`} {
		e, f := assertionEnv(t, `probe(curl_setopt(`+args+`));`, false, phpversion.Max)
		if _, known := e.BuiltinArguments(assertionRead(f).(*syntax.FuncCall)); known {
			t.Fatal("pack budget ignored")
		}
	}
}

func BenchmarkNestedArgumentPacks(b *testing.B) {
	file := syntax.Parse("bench.php", []byte(`<?php curl_setopt(...[...['handle'=>$h],...['option'=>CURLOPT_SSL_VERIFYPEER,'value'=>false]]);`), syntax.Options{Version: phpversion.Max})
	e := NewEnv(file, names.New(file), index.New(stubs.Index()), phpversion.Max, Options{})
	call := file.Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.FuncCall)
	b.ReportAllocs()
	for b.Loop() {
		e.BuiltinArguments(call)
	}
}
