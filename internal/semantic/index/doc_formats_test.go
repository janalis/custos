package index

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"custos/internal/php/phpdoc"
)

func TestTemplateVarianceIsRecorded(t *testing.T) {
	fs := extract(t, "variance.php", `<?php
/**
 * @template A
 * @template-covariant B
 * @psalm-template-contravariant C
 * @phpstan-template-contravariant D
 */
class Box {}
`)
	var got []phpdoc.Variance
	for _, tpl := range fs.Classes[0].Templates {
		got = append(got, tpl.Variance)
	}
	want := []phpdoc.Variance{phpdoc.Invariant, phpdoc.Covariant, phpdoc.Contravariant, phpdoc.Contravariant}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("variances = %v, want %v", got, want)
	}
}

func TestPrefixedMagicMembersWin(t *testing.T) {
	fs := extract(t, "magic_prefixed.php", `<?php
/**
 * @property int $a
 * @phpstan-property string $a
 * @psalm-property bool $a
 * @psalm-property bool $b
 * @property int $b
 * @psalm-property-read float $c
 * @phpstan-property-write array $d
 * @method int m()
 * @phpstan-method string m()
 * @psalm-method bool n()
 * @method int n()
 * @phpstan-method not a signature(
 */
class Magic {}
`)
	c := fs.Classes[0]
	for name, want := range map[string]string{"a": "string", "b": "bool", "c": "float", "d": "array"} {
		if p := c.Props[name]; p == nil || p.DocType != want || !p.Magic {
			t.Errorf("$%s = %+v, want %s", name, p, want)
		}
	}
	if !c.Props["c"].Readonly || c.Props["d"].Readonly {
		t.Error("read-only flags")
	}
	for name, want := range map[string]string{"m": "string", "n": "bool"} {
		if m := c.Methods[name]; m == nil || m.DocReturn != want {
			t.Errorf("%s() = %+v, want %s", name, m, want)
		}
	}
}

func TestMixinClasses(t *testing.T) {
	fs := extract(t, "mixins.php", `<?php
namespace App;
use Domain\Item;
/**
 * @template T
 * @mixin Item
 * @mixin Item
 * @phpstan-mixin \Other\Thing<int>
 * @psalm-mixin T
 */
class Host {}
/** @mixin */
class Bare {}
class Plain {}
`)
	host := fs.Classes[0]
	if !host.Mixin || !reflect.DeepEqual(host.Mixins, []string{`Domain\Item`, `Other\Thing`}) {
		t.Errorf("Host: mixin=%v mixins=%v", host.Mixin, host.Mixins)
	}
	if c := fs.Classes[1]; !c.Mixin || c.Mixins != nil {
		t.Errorf("bare @mixin: mixin=%v mixins=%v", c.Mixin, c.Mixins)
	}
	if c := fs.Classes[2]; c.Mixin || c.Mixins != nil {
		t.Errorf("no tag: mixin=%v mixins=%v", c.Mixin, c.Mixins)
	}
}

func TestSelfOut(t *testing.T) {
	fs := extract(t, "self_out.php", `<?php
namespace App;
/** @template T */
class Box {
	/** @phpstan-self-out self<int> */
	function a() {}
	/** @psalm-this-out Box<T> */
	function b() {}
	/** @this-out static */
	function c() {}
	/**
	 * @self-out Box<string>
	 * @phpstan-this-out Box<int>
	 */
	function d() {}
	/** @return int */
	function e() {}
	/** @psalm-self-out $this<string> */
	function f() {}
}
`)
	ms := fs.Classes[0].Methods
	for name, want := range map[string]string{
		"a": `\App\Box<int>`, "b": `\App\Box<\~T>`, "c": "static", "d": `\App\Box<int>`, "e": "", "f": `\App\Box<string>`,
	} {
		if got := ms[name].SelfOut; got != want {
			t.Errorf("%s: SelfOut = %q, want %q", name, got, want)
		}
	}
}

func TestMixinsAreCapped(t *testing.T) {
	var doc strings.Builder
	for i := range maxClassMixins + 5 {
		fmt.Fprintf(&doc, " * @mixin M%d\n", i)
	}
	fs := extract(t, "many.php", "<?php\n/**\n"+doc.String()+" */\nclass Many {}\n")
	if c := fs.Classes[0]; !c.Mixin || len(c.Mixins) != maxClassMixins {
		t.Errorf("mixin=%v, %d mixins kept", c.Mixin, len(c.Mixins))
	}
}
