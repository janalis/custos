package infer_test

import "testing"

// Loose comparisons with booleans test truthiness; loose inequality with a
// falsy constant excludes the null and false it equals.
func TestLooseComparisonNarrowing(t *testing.T) {
	checkWith(t, nil, `<?php
class Obj {}
function run(string|false $s, ?Obj $o, string|false|null $n, bool $b, int|false $i) {
    if ($s == false) { t('eqFalse', $s); } else { t('neFalseElse', $s); }
    if ($s != false) { t('neFalse', $s); }
    if (false != $o) { t('neFalseFlipped', $o); }
    if ($b == true) { t('eqTrue', $b); } else { t('eqTrueElse', $b); }
    if ($b != true) { t('neTrue', $b); }
    if ($n != null) { t('neNull', $n); }
    if ($n != '') { t('neEmpty', $n); }
    if ($n != '0') { t('neZeroString', $n); }
    if ($i != 0) { t('neZero', $i); }
    if ($i != 0.0) { t('neZeroFloat', $i); }
    if ($i != 1) { t('neOne', $i); }
    if ($n != 'x') { t('neOther', $n); }
    if ($n == '') { t('eqEmpty', $n); }
    if ($i != 0x0) { t('hex', $i); }
}
`, map[string]string{
		"eqFalse": "false|string", "neFalseElse": "string", "neFalse": "string", "neFalseFlipped": `\Obj`,
		"eqTrue": "true", "eqTrueElse": "false", "neTrue": "false",
		"neNull": "string", "neEmpty": "string", "neZeroString": "null|string",
		"neZero": "int", "neZeroFloat": "int", "neOne": "false|int", "neOther": "false|null|string",
		"eqEmpty": "false|null|string", "hex": "false|int",
	})
}
