package index

import (
	"reflect"
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
