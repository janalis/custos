package infer_test

import "testing"

// Integer literals carry their value and string literals their byte length;
// unary minus negates the value, and operators build plain types.
func TestLiteralValueFacts(t *testing.T) {
	checkShape(t, `<?php
/**
 * @param int<1, 5> $small
 * @param int<min, 0> $low
 */
function g($small, $low) {
    t('negRange', -$small);
    t('negMin', -$low);
}
function f(int $n) {
    t('int', 42);
    t('hex', 0x1F);
    t('octal', 0o17);
    t('sep', 1_000);
    t('neg', -1);
    t('plus', +3);
    t('negNeg', -(-2));
    t('negVar', -$n);
    t('min', -9223372036854775807 - 1);
    t('str', 'abc');
    t('empty', '');
    t('escaped', "a\n");
    t('heredoc', <<<EOT
    xy
    EOT);
    t('nowdoc', <<<'EOT'
    xyz
    EOT);
    t('interp', "a$n");
    t('sum', 1 + 2);
    t('concat', 'a' . 'b');
    t('float', 1.5);
    t('badEscape', "\u{110000}");
    t('ternary', $n ? 1 : 2);
    $x = 5;
    t('var', $x);
}
`, map[string]string{
		"int": "int<42, 42>", "hex": "int<31, 31>", "octal": "int<15, 15>", "sep": "int<1000, 1000>",
		"neg": "int<-1, -1>", "plus": "int<3, 3>", "negNeg": "int<2, 2>", "negVar": "int",
		"min": "int", "negRange": "int<-5, -1>", "negMin": "int",
		"str": "string<len 3, 3>", "empty": "string<len 0, 0>", "escaped": "string<len 2, 2>",
		"heredoc": "string<len 2, 2>", "nowdoc": "string<len 3, 3>",
		"interp": "string", "sum": "int", "concat": "string", "float": "float", "badEscape": "string",
		"ternary": "int<1, 2>", "var": "int<5, 5>",
	})
}
