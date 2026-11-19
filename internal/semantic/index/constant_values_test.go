package index

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func constantIndex(t *testing.T, src string) *Index {
	t.Helper()
	ix := New(nil)
	ix.Add(Extract(syntax.Parse("c.php", []byte(src), syntax.Options{Version: phpversion.PHP84})))
	return ix
}

func TestConstantValues(t *testing.T) {
	ix := constantIndex(t, `<?php
namespace App;
const LOCAL = 3;
class Base { const A = 1; const B_X = 'x'; }
class Child extends Base { const A = 1; const B_Y = 'y'; const C = 2; }
class Odd { const A = self::B; const B = 1; }
class Orphan extends Missing { const A = 1; }
interface Flags { const F1 = 1; }
class Flagged implements Flags {}
class NoConsts {}
`)
	for expr, want := range map[string]string{
		`\App\Base::A`:       "int<1, 1>",
		`\App\Child::B_*`:    "'x'|'y'",
		`\App\Child::*`:      "int<1, 2>|'x'|'y'",
		`\App\Flagged::*`:    "int<1, 1>",
		`\App\Odd::A`:        "",
		`\App\Odd::*`:        "",
		`\App\Orphan::*`:     "",
		`\App\NoConsts::*`:   "",
		`\App\Base::NONE`:    "",
		`\::App\LOCAL`:       "int<3, 3>",
		`\::App\E_NOWHERE`:   "",
		`\::App\MISSING_ONE`: "",
	} {
		got := ConstantValues(ix, expr, phpversion.PHP84)
		text := got.ValueDocString()
		if got.IsUnknown() {
			text = ""
		}
		if text != want {
			t.Errorf("%s = %q, want %q", expr, text, want)
		}
	}
}

// A hierarchy cut at the ancestor cap may hide constants.
func TestConstantValuesCutHierarchy(t *testing.T) {
	var src strings.Builder
	src.WriteString("<?php\n")
	var names []string
	for i := range MaxAncestors + 4 {
		fmt.Fprintf(&src, "interface I%d {}\n", i)
		names = append(names, fmt.Sprintf("I%d", i))
	}
	fmt.Fprintf(&src, "class Wide implements %s { const A = 1; }\n", strings.Join(names, ", "))
	ix := constantIndex(t, src.String())
	if got := ConstantValues(ix, `\Wide::*`, phpversion.PHP84); !got.IsUnknown() {
		t.Errorf("cut hierarchy = %s", got.ValueDocString())
	}
}

func TestLiteralType(t *testing.T) {
	for text, want := range map[string]string{
		"1": "int<1, 1>", " 0x10 ": "int<16, 16>", "'a'": "'a'", `"b"`: "'b'",
		`'a\'b'`: "", `"$x"`: "", "1.5": "", "FOO": "", "'": "",
	} {
		got := LiteralType(text)
		text2 := got.ValueDocString()
		if got.IsUnknown() {
			text2 = ""
		}
		if text2 != want {
			t.Errorf("LiteralType(%q) = %q, want %q", text, text2, want)
		}
	}
}
