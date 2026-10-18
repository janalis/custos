package index

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestDecodedAndComposedDeclarationNames(t *testing.T) {
	source := `<?php namespace Studio;
    class Original {}
    define('Studio\\First', 1);
    define("Studio\x5cSecond", 2);
    define((__NAMESPACE__ . '\\' . 'Third'), 3);
    define(Original::class . '\\Fourth', 4);
    define('\\Studio\\Hidden', 5);
    define('', 6);
    class_alias(Original::class, __NAMESPACE__ . '\\' . 'Alias');
    class_alias('Studio\\Original', '\\Studio\\EscapedAlias');
    `
	file := syntax.Parse("names.php", []byte(source), syntax.Options{})
	if len(file.Errors) != 0 {
		t.Fatal(file.Errors)
	}
	symbols := Extract(file)
	ix := New(nil)
	ix.Add(symbols)
	for _, name := range []string{`Studio\First`, `Studio\Second`, `Studio\Third`, `Studio\Original\Fourth`, ""} {
		if ix.Constant(name, phpversion.PHP85) == nil {
			t.Fatalf("missing constant %q: %+v", name, symbols.Constants)
		}
	}
	if ix.Constant(`Studio\Hidden`, phpversion.PHP85) != nil || ix.Constant(`\Studio\Hidden`, phpversion.PHP85) != nil {
		t.Fatal("leading separator was stripped from define name")
	}
	if ix.Constant(`\\Studio\Hidden`, phpversion.PHP85) == nil {
		t.Fatal("literal define name was not retained")
	}
	for _, name := range []string{`Studio\Alias`, `Studio\EscapedAlias`} {
		if class := ix.Class(name, phpversion.PHP85); class == nil || class.FQN != `Studio\Original` {
			t.Fatalf("alias %s: %+v", name, class)
		}
	}
	ix.Remove("names.php")
	if ix.Constant(`\\Studio\Hidden`, phpversion.PHP85) != nil {
		t.Fatal("literal constant survived removal")
	}
}

func TestUnknownAndBoundedDeclarationNames(t *testing.T) {
	symbols := extract(t, "unknown.php", `<?php define($name,1);define('X'.unknown(),1);define('X' . 7,1);define('A'+ 'B',1);define(__FILE__,1);class_alias(self::class,'Alias');class_alias('Original','');`)
	if len(symbols.Constants) != 0 || len(symbols.ClassAliases) != 0 {
		t.Fatal("unknown expression invented a name")
	}
	source := `<?php define(` + strings.Repeat(`'part'.`, 300) + `'end',1);`
	if got := extract(t, "bounded.php", source); len(got.Constants) != 0 {
		t.Fatal("declaration expression exceeded its bound")
	}
}

func TestDeclarationNameVersion(t *testing.T) {
	for _, test := range []struct {
		version phpversion.Version
		want    string
	}{
		{phpversion.PHP53, `Prefix\u{41}\e`},
		{phpversion.PHP54, "Prefix\\u{41}\x1b"},
		{phpversion.PHP70, "PrefixA\x1b"},
	} {
		file := syntax.Parse("version.php", []byte(`<?php define("Prefix\u{41}\e",1);`), syntax.Options{Version: test.version})
		symbols := Extract(file)
		if len(symbols.Constants) != 1 || symbols.Constants[0].FQN != test.want {
			t.Fatalf("PHP %v: %+v want %q", test.version, symbols.Constants, test.want)
		}
	}
}

func BenchmarkComposedDeclarationName(b *testing.B) {
	file := syntax.Parse("names.php", []byte(`<?php namespace Studio;define(__NAMESPACE__.'\\'.'Entry',1);`), syntax.Options{})
	b.ReportAllocs()
	for b.Loop() {
		Extract(file)
	}
}
