package flow

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

func TestLiteralTextPHPEscapes(t *testing.T) {
	for _, tc := range []struct {
		raw, want string
		version   phpversion.Version
	}{
		{`"\a"`, `\a`, phpversion.PHP85},
		{`"\b"`, `\b`, phpversion.PHP85},
		{`"\u0030"`, `\u0030`, phpversion.PHP85},
		{`"\q"`, `\q`, phpversion.PHP85},
		{`"\x5fGET"`, `_GET`, phpversion.PHP85},
		{"<<<TEXT\n\\q\nTEXT\n", `\q`, phpversion.PHP85},
		{"<<<'TEXT'\n\\e\nTEXT\n", `\e`, phpversion.PHP85},
		{`b"name"`, `name`, phpversion.PHP85},
		{`"\\e"`, `\e`, phpversion.PHP53},
		{`"\e"`, `\e`, phpversion.PHP53},
		{`"\e"`, "\x1b", phpversion.PHP54},
		{`"\u{30}"`, `\u{30}`, phpversion.PHP53},
		{`"\u{30}"`, `0`, phpversion.PHP70},
	} {
		t.Run(tc.raw+tc.version.String(), func(t *testing.T) {
			file := syntax.Parse("literal.php", []byte("<?php probe("+tc.raw+");"), syntax.Options{Version: tc.version})
			if len(file.Errors) > 0 {
				t.Fatal(file.Errors)
			}
			call := file.Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.FuncCall)
			value := call.Args.Args[0].(*syntax.Arg).Value
			if got, known := literalText(value); !known || got != tc.want {
				t.Fatalf("value %q known=%v want %q", got, known, tc.want)
			}
		})
	}
}

func TestConstantStringRetainsQuotedSubset(t *testing.T) {
	for _, raw := range []string{`b"text"`, "<<<TEXT\ntext\nTEXT", ""} {
		if _, known := constantString(&syntax.Literal{LitKind: syntax.LitString, Raw: raw}); known {
			t.Fatalf("unexpected folded form %q", raw)
		}
	}
}

func BenchmarkFlowLiteralText(b *testing.B) {
	for _, raw := range []string{`"plain"`, `"\a"`, `"\\r\\n"`} {
		b.Run(raw, func(b *testing.B) {
			file := syntax.Parse("literal.php", []byte("<?php probe("+raw+");"), syntax.Options{})
			value := file.Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.FuncCall).Args.Args[0].(*syntax.Arg).Value
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				literalText(value)
			}
		})
	}
}

func TestPHPStringKeysKeepIndependentTaint(t *testing.T) {
	for _, tc := range []struct {
		source  string
		tainted bool
	}{
		{`$values=["\a"=>$_GET['x'],"\x07"=>'safe'];probe($values["\a"]);`, true},
		{`$values=["\a"=>'safe',"\x07"=>$_GET['x']];probe($values["\a"]);`, false},
		{`$values=["\u0030"=>$_GET['x'],0=>'safe'];probe($values["\u0030"]);`, true},
		{`$values=["\u0030"=>'safe',0=>$_GET['x']];probe($values["\u0030"]);`, false},
		{`$x=${b"_GET"}['x'];probe($x);`, true},
		{`$x=${"\x5fGET"}['x'];probe($x);`, true},
		{`$x=${"\u005fGET"}['x'];probe($x);`, false},
	} {
		t.Run(tc.source, func(t *testing.T) {
			env, file := environment(t, tc.source, nil)
			value := argument(callsNamed(file, "probe")[0])
			if got := env.Tainted(value, HTML); got != tc.tainted {
				t.Fatalf("tainted=%v want=%v value=%+v", got, tc.tainted, env.Value(value))
			}
		})
	}
}

func TestParsedUnicodeStringKeyVersions(t *testing.T) {
	for _, version := range []phpversion.Version{phpversion.PHP53, phpversion.PHP70, phpversion.PHP85} {
		t.Run(version.String(), func(t *testing.T) {
			file := syntax.Parse("literal.php", []byte(`<?php $values=["\u{30}"=>$_GET['x'],0=>'safe'];probe($values["\u{30}"]);`), syntax.Options{Version: version})
			if len(file.Errors) > 0 {
				t.Fatal(file.Errors)
			}
			ix := index.New(stubs.Index())
			ix.Add(index.Extract(file))
			env := New(file, infer.NewEnv(file, names.New(file), ix, version), nil)
			value := argument(callsNamed(file, "probe")[0])
			if got, want := env.Tainted(value, HTML), version < phpversion.PHP70; got != want {
				t.Fatalf("tainted=%v want=%v", got, want)
			}
		})
	}
}
