package infer_test

import (
	"strings"
	"testing"
)

// Array literal constant initialisers keep their shape, element type and
// emptiness; values the extractor cannot type are mixed.
func TestConstantArrayShapes(t *testing.T) {
	checkShape(t, `<?php
namespace App;
const NAMES = ['a', 'b'];
const NONE = [];
define('App\DEFINED', ['x' => 1.5]);
class Foo {
    const SIZE = 'size';
    const MAP = ['name' => 'n', self::SIZE => 3, Foo::NESTED => true, 'neg' => -1, 'pos' => +2.5, 'str' => -'x'];
    const NESTED = 'nested';
    const DEEP = ['row' => ['id' => 1, 'tags' => ['a']], 'line' => __LINE__, 'file' => __FILE__, 'null' => null, 'heredoc' => <<<TXT
    text
    TXT];
    const REF = [self::LIST_REF, Other::X, parent::Y, self::MISSING];
    const LIST_REF = [1];
    const KEYED = [Other::KEY => 1, 2 => 'x'];
    const SPREAD = [...self::LIST_REF, 'b' => 'c'];
    const AUTO_AFTER_NEGATIVE = [-5 => 'a', 'b'];
    const DUPLICATE = ['a' => 1, 'a' => 'x'];
    const CYCLE = [self::CYCLE_B => 1];
    const CYCLE_B = self::CYCLE_C;
    const CYCLE_C = self::CYCLE_B;
    const ONLY_MIXED = [Other::KEY => Other::VALUE];
    const ALL_NEVER = [Other::KEY => []];
    const SCALAR = 'x';
    const PAREN = (['p' => 1]);
    final const FINAL = ['f' => 1];
    public function f() {
        t('map', self::MAP); t('deep', self::DEEP); t('ref', self::REF);
        t('keyed', self::KEYED); t('spread', self::SPREAD); t('autoNeg', self::AUTO_AFTER_NEGATIVE);
        t('dup', self::DUPLICATE); t('cycle', self::CYCLE); t('onlyMixed', self::ONLY_MIXED);
        t('allNever', self::ALL_NEVER); t('scalar', self::SCALAR); t('paren', self::PAREN);
        t('named', Foo::MAP['name']); t('static', static::MAP); t('staticFinal', static::FINAL);
        t('object', $this::MAP); t('global', NAMES); t('empty', NONE); t('defined', DEFINED);
    }
}
final class Sealed { const S = ['s' => 1]; public function f() { t('finalClass', static::S); } }
enum E { const C = ['c' => 1]; public function f() { t('enum', static::C); } }
`, map[string]string{
		"map":         "array{name: string, size: int, nested: true, neg: int, pos: float, str: mixed}",
		"deep":        "array{row: array{id: int, tags: array{0: string}}, line: int, file: string, null: null, heredoc: string}",
		"ref":         "array{0: array{0: int}, 1: mixed, 2: mixed, 3: mixed}",
		"keyed":       "non-empty int[]|string[]",
		"spread":      "non-empty mixed[]|string[]",
		"autoNeg":     "non-empty string[]",
		"dup":         "array{a: string}",
		"cycle":       "non-empty int[]",
		"onlyMixed":   "non-empty mixed[]",
		"allNever":    "non-empty array[]",
		"scalar":      "string",
		"paren":       "array{p: int}",
		"named":       "string",
		"static":      "array",
		"staticFinal": "array{f: int}",
		"object":      "array",
		"global":      "array{0: string, 1: string}",
		"empty":       "array{}",
		"defined":     "array{x: float}",
		"finalClass":  "array{s: int}",
		"enum":        "array{c: int}",
	})
}

// A constant with more keys than a shape holds keeps its element type.
func TestConstantArrayOverShapeCap(t *testing.T) {
	var b strings.Builder
	for i := range 40 {
		b.WriteString("'k")
		b.WriteByte(byte('a' + i%26))
		b.WriteString(strings.Repeat("x", i/26))
		b.WriteString("' => 1, ")
	}
	checkShape(t, `<?php
class Big { const B = [`+b.String()+`]; public function f() { t('big', self::B); foreach (self::B as $v) { t('v', $v); } } }
`, map[string]string{"big": "non-empty int[]", "v": "int"})
}
