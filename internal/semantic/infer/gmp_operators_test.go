package infer_test

import "testing"

func TestOperatorOverloadedObjects(t *testing.T) {
	check(t, `<?php
function f(\GMP $g, \BcMath\Number $n, int $i, $u) {
    t('add', $g + 1); t('sub', $n - $i); t('mul', $i * $g); t('div', $g / 2); t('pow', $n ** 2);
    t('mod', $g % 2); t('shl', $g << 1); t('shr', $i >> $g); t('nmod', $n % 3);
    t('imod', $i % 2); t('umod', $u % 2);
    $a = $g; $a %= 2; t('modeq', $a);
    $b = $g; $b <<= 1; t('shleq', $b);
    $c = $i; $c %= 2; t('imodeq', $c);
}
`, map[string]string{
		"add": `\GMP`, "sub": `\BcMath\Number`, "mul": `\GMP`, "div": `\GMP`, "pow": `\BcMath\Number`,
		"mod": `\GMP`, "shl": `\GMP`, "shr": `\GMP`, "nmod": `\BcMath\Number`,
		"imod": "int", "umod": "int", "modeq": `\GMP`, "shleq": `\GMP`, "imodeq": "int",
	})
}

// `**` on ints is an int only when literal operands prove a non-negative
// exponent and no overflow: `2 ** -1` is 0.5, `2 ** 63` overflows to a float.
func TestPowerOfInts(t *testing.T) {
	check(t, `<?php
function f(int $i, int $j) {
    t('lit', 2 ** 10); t('neg', 2 ** -1); t('over', 2 ** 63); t('max', 2 ** 62);
    t('vars', $i ** $j); t('varexp', 2 ** $j); t('varbase', $i ** 2);
    $a = 2; $a **= -1; t('poweq', $a);
}
`, map[string]string{
		"lit": "int", "neg": "float|int", "over": "float|int", "max": "int",
		"vars": "float|int", "varexp": "float|int", "varbase": "float|int", "poweq": "float|int",
	})
}
