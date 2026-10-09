package infer

import (
	"fmt"
	"testing"

	"custos/internal/index"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

func TestResolvedConstants(t *testing.T) {
	cases := []struct{ src, want string }{
		{`namespace App; const PHP_INT_MAX = 'local'; PHP_INT_MAX;`, "string"},
		{`namespace App; const PHP_VERSION = 12; PHP_VERSION;`, "int"},
		{`const PHP_VERSION = 12; PHP_VERSION;`, "string"},
		{`const M_PI = "other"; M_PI;`, "float"},
		{`namespace App; const PHP_INT_MAX = 'local'; \PHP_INT_MAX;`, "int"},
		{`namespace App; PHP_INT_MAX;`, "int"},
		{`namespace App; namespace\PHP_INT_MAX;`, "?unknown"},
		{`namespace App; Other\PHP_INT_MAX;`, "?unknown"},
		{`namespace Lib { const size = 'small'; } namespace App { use const Lib\size as PHP_INT_MAX; PHP_INT_MAX; }`, "string"},
		{`namespace App; use const Missing\size as PHP_VERSION; PHP_VERSION;`, "?unknown"},
		{`namespace App; use const PHP_INT_MAX as largest; largest;`, "int"},
		{`php_int_max;`, "?unknown"},
		{`\php_version;`, "?unknown"},
		{`const SIZE = 1; size;`, "?unknown"},
		{`namespace App; TrUe;`, "true"},
		{`namespace App; \FaLsE;`, "false"},
		{`namespace App; NuLl;`, "null"},
		{`namespace App; Missing;`, "?unknown"},
		{`const SIZE = other(); SIZE;`, "?unknown"},
		{`namespace App; const PHP_INT_MAX = other(); PHP_INT_MAX;`, "?unknown"},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			f := syntax.Parse("constants.php", []byte("<?php "+tc.src), syntax.Options{Version: phpver.PHP85})
			if len(f.Errors) != 0 {
				t.Fatal(f.Errors)
			}
			ix := index.New(stubs.Index())
			ix.Add(index.Extract(f))
			env := NewEnv(f, names.New(f), ix, phpver.PHP85)
			var expr syntax.Expr
			syntax.InspectFile(f, func(n syntax.Node) bool {
				if s, ok := n.(*syntax.ExprStmt); ok {
					expr = s.Expr
				}
				return true
			})
			for _, e := range []*Env{env, env.Native()} {
				if got := e.TypeOf(expr).String(); got != tc.want {
					t.Errorf("native=%v: got %s want %s", e.IsNative(), got, tc.want)
				}
			}
		})
	}
}

func TestConstantAvailability(t *testing.T) {
	f := syntax.Parse("availability.php", []byte(`<?php namespace App; Future;`), syntax.Options{Version: phpver.PHP85})
	ix := index.New(nil)
	ix.Add(&index.FileSymbols{Path: "constants.php", Constants: []*index.Constant{
		{FQN: `App\Future`, Value: "1", Avail: index.Avail{From: phpver.PHP85}},
		{FQN: "Future", Value: "'fallback'"},
	}})
	env := NewEnv(f, names.New(f), ix, phpver.PHP84)
	var expr syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if s, ok := n.(*syntax.ExprStmt); ok {
			expr = s.Expr
		}
		return true
	})
	for _, e := range []*Env{env, env.Native()} {
		if got := e.TypeOf(expr).String(); got != "string" {
			t.Errorf("fallback: %s", got)
		}
	}
	for _, tc := range []struct {
		name string
		ver  phpver.Version
		want string
	}{
		{"PHP_INT_MIN", phpver.PHP56, "?unknown"},
		{"PHP_INT_MIN", phpver.PHP70, "int"},
		{"PHP_FLOAT_EPSILON", phpver.PHP71, "?unknown"},
		{"PHP_FLOAT_EPSILON", phpver.PHP72, "float"},
		{"PHP_OS_FAMILY", phpver.PHP71, "?unknown"},
		{"PHP_OS_FAMILY", phpver.PHP72, "string"},
	} {
		env := NewEnv(f, names.New(f), stubs.Index(), tc.ver)
		got, _ := env.namedConstType(tc.name)
		if got.String() != tc.want {
			t.Errorf("%s at %s: got %s want %s", tc.name, tc.ver, got, tc.want)
		}
	}
}

func TestConstantPolyfill(t *testing.T) {
	f := syntax.Parse("polyfill.php", []byte(`<?php const PHP_INT_MIN = 'polyfill'; PHP_INT_MIN;`), syntax.Options{Version: phpver.PHP56})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := NewEnv(f, names.New(f), ix, phpver.PHP56)
	expr := f.Stmts[len(f.Stmts)-1].(*syntax.ExprStmt).Expr
	for _, e := range []*Env{env, env.Native()} {
		if got := e.TypeOf(expr).String(); got != "string" {
			t.Errorf("polyfill: %s", got)
		}
	}
}

func TestBuiltinConstantTypes(t *testing.T) {
	for _, tc := range []struct {
		name string
		want string
	}{
		{"PHP_INT_MIN", "int"},
		{"PHP_INT_MAX", "int"},
		{"PHP_OS_FAMILY", "string"},
		{"PHP_FLOAT_MIN", "float"},
		{"PHP_VERSION", "string"},
		{"M_PI", "float"},
		{"missing", "?unknown"},
	} {
		for _, ver := range []phpver.Version{0, phpver.PHP85} {
			if got := builtinConstType(tc.name, ver).String(); got != tc.want {
				t.Errorf("%s at %s: got %s want %s", tc.name, ver, got, tc.want)
			}
		}
	}
}

func TestConstantsWithoutStubInitializers(t *testing.T) {
	f := syntax.Parse("empty-stubs.php", []byte("<?php"), syntax.Options{Version: phpver.PHP85})
	ix := index.New(nil)
	ix.Add(&index.FileSymbols{Path: "stub.php", Constants: []*index.Constant{{FQN: "PHP_VERSION"}}})
	env := NewEnv(f, names.New(f), ix, phpver.PHP85)
	for _, tc := range []struct{ name, want string }{{"PHP_VERSION", "string"}, {"M_PI", "float"}} {
		got, found := env.namedConstType(tc.name)
		if !found || got.String() != tc.want {
			t.Errorf("%s: found=%v got %s want %s", tc.name, found, got, tc.want)
		}
	}
	if got := builtinConstType("PHP_INT_MIN", phpver.PHP56); !got.IsUnknown() {
		t.Errorf("unavailable builtin: %s", got)
	}
}

func BenchmarkConstantResolution(b *testing.B) {
	f := syntax.Parse("bench.php", []byte(`<?php namespace App; PHP_INT_MAX;`), syntax.Options{Version: phpver.PHP85})
	env := NewEnv(f, names.New(f), stubs.Index(), phpver.PHP85)
	var expr *syntax.ConstFetch
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.ConstFetch); ok {
			expr = c
		}
		return true
	})
	b.ReportAllocs()
	for b.Loop() {
		_ = env.resolvedConstType(expr)
	}
}
