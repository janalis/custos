package index

import (
	"testing"

	phpversion "custos/internal/php/version"
)

// Stub return types vary by PHP version (#[LanguageLevelTypeAware]): the
// index resolves them per requested version, returning one stable copy per
// declaration and phpversion.
func TestVersionedReturnTypes(t *testing.T) {
	fs := extract(t, "stubs.php", `<?php
#[LanguageLevelTypeAware(["8.0" => "string"], default: "string|false")]
function substr($s) {}
#[LanguageLevelTypeAware(["7.1" => "int", "8.0" => "int|float"], default: "")]
function steps() {}
#[LanguageLevelTypeAware(["8.0" => "int"], default: "int")]
function same() {}
#[LanguageLevelTypeAware(["7.0" => "int", "8.0" => "int"], default: "bool")]
function twice() {}
#[LanguageLevelTypeAware([], default: "bool")]
function onlyDefault() {}
#[LanguageLevelTypeAware(["x" => "int", 5], default: "")]
function badKeys() {}
#[LanguageLevelTypeAware("int")]
function notAMap() {}
#[AllowDynamicProperties, \Other\Attr]
class C {
    #[LanguageLevelTypeAware(["8.1" => "mixed"], default: "array|false")]
    public function get() {}
}
`)
	ix := New(nil)
	ix.Add(fs)
	cases := []struct {
		fn   string
		ver  phpversion.Version
		want string
	}{
		{"substr", phpversion.PHP74, "false|string"},
		{"substr", phpversion.PHP80, "string"},
		{"substr", 0, "string"},
		{"steps", phpversion.PHP70, ""},
		{"steps", phpversion.PHP74, "int"},
		{"steps", phpversion.PHP84, "float|int"},
		{"same", phpversion.PHP74, "int"},
		{"onlyDefault", phpversion.PHP74, "bool"},
		{"badKeys", phpversion.PHP74, ""},
		{"notAMap", phpversion.PHP74, ""},
		{"twice", phpversion.PHP56, "bool"},
		{"twice", phpversion.PHP74, "int"},
	}
	for _, c := range cases {
		if got := ix.Function(c.fn, c.ver).Return; got != c.want {
			t.Errorf("%s at %v: %q, want %q", c.fn, c.ver, got, c.want)
		}
	}
	if ix.Function("same", 0).RetVer != nil {
		t.Error("a type equal at every version needs no version map")
	}
	if a, b := ix.Function("substr", phpversion.PHP74), ix.Function("substr", phpversion.PHP74); a != b {
		t.Error("the version copy must be stable")
	}
	if ix.Function("substr", phpversion.PHP80) != ix.Function("substr", 0) {
		t.Error("the newest version returns the declaration itself")
	}
	m74, m84 := ix.FindMethod("C", "get", phpversion.PHP74), ix.FindMethod("C", "get", phpversion.PHP84)
	if again := ix.FindMethod("C", "get", phpversion.PHP74); m74.Return != "array|false" || m84.Return != "mixed" || m74 != again {
		t.Errorf("method: %q / %q", m74.Return, m84.Return)
	}
	c := ix.Class("C", 0)
	if !c.HasAttr(`\AllowDynamicProperties`) || !c.HasAttr("other\\attr") || c.HasAttr("Missing") {
		t.Errorf("class attributes: %v", c.Attrs)
	}
}

func TestFunctionDecls(t *testing.T) {
	base := New(nil)
	base.Add(extract(t, "stub.php", `<?php function strlen($s) {}`))
	ix := New(base)
	ix.Add(extract(t, "a.php", `<?php function dup() {} function strlen($s) {}`))
	ix.Add(extract(t, "b.php", `<?php function dup() {}`))
	if got := ix.FunctionDecls("dup", phpversion.PHP84); len(got) != 2 {
		t.Errorf("dup: %d declarations", len(got))
	}
	if got := ix.FunctionDecls("strlen", phpversion.PHP84); got != nil {
		t.Errorf("a polyfill of a builtin: %v", got)
	}
	if got := ix.FunctionDecls("nope", phpversion.PHP84); got != nil {
		t.Errorf("undeclared: %v", got)
	}
	if got := New(nil).FunctionDecls("nope", 0); got != nil {
		t.Errorf("no base: %v", got)
	}
}
