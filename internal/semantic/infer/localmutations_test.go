package infer_test

import "testing"

func TestLocalUnset(t *testing.T) {
	check(t, `<?php
function unsetValues(bool $condition) {
    $x = 's'; unset($x); t('unset', $x);
    $x = 1; t('restored', $x);
    $y = 's'; if ($condition) { unset($y); } t('conditional', $y);
    $unbraced = 's'; if ($condition) unset($unbraced); t('unbraced', $unbraced);
    $z = 's'; if ($condition) { unset($z); } else { $z = 1; } t('branches', $z);
    $array = ['x' => 1]; unset($array['x']); t('element', $array);
    $one = 1; $two = 's'; unset($one, $two); t('one', $one); t('two', $two);
    $param = 's'; while ($condition) { t('loop', $param); unset($param); }
}
`, map[string]string{
		"unset": "null", "restored": "int", "conditional": "null|string", "unbraced": "null|string", "branches": "int|null|string",
		"element": "int[]", "one": "null", "two": "null", "loop": "null|string",
	})
}

func TestLocalIncrementDecrement(t *testing.T) {
	check(t, `<?php
function mutations(bool $condition, int $integer, float $float, string $text, $unknown) {
    $x = null; t('prefix', ++$x); t('prefixStored', $x);
    $y = null; t('postfix', $y++); t('postfixStored', $y);
    $z = null; t('decrement', --$z); t('decrementStored', $z);
    $a = null; $result = ++$a; t('nested', $a);
    $b = null; $condition && ++$b; t('optional', $b);
    $c = null; if (++$c && $c) { t('condition', $c); }
    t('integer', ++$integer); t('float', --$float); t('text', ++$text); t('unknown', ++$unknown);
    $true = true; t('true', ++$true);
    $false = false; t('false', --$false);
    $bool = $condition; t('bool', ++$bool);
    $array = []; t('invalid', ++$array);
    $mixed = $condition ? null : 1.0; t('union', ++$mixed);
    $reset = null; ++$reset; $reset = 's'; t('reassigned', $reset);
}
`, map[string]string{
		"prefix": "int", "prefixStored": "int", "postfix": "null", "postfixStored": "int",
		"decrement": "null", "decrementStored": "null", "nested": "int", "optional": "int|null",
		"condition": "int", "integer": "float|int", "float": "float", "text": "float|int|string",
		"unknown": "?unknown", "true": "true", "false": "false", "bool": "bool", "invalid": "?unknown",
		"union": "float|int", "reassigned": "string",
	})
}

func TestMutationOperandOrderAndBoundaries(t *testing.T) {
	checkAnywhere(t, `<?php
function ordered(array $a) {
    $key = 's'; unset($key, $a[t('laterUnset', $key)]);
    $x = null; ++$x && t('laterIncrement', $x);
    $max = PHP_INT_MAX; ++$max; t('max', $max);
    $min = -PHP_INT_MAX - 1; --$min; t('min', $min);
    $integerString = '99'; ++$integerString; t('integerString', $integerString);
    $decimalString = '1.5'; --$decimalString; t('decimalString', $decimalString);
}

$top = null; ++$top; t('topIncrement', $top);
unset($top); t('topUnset', $top);
`, map[string]string{
		"laterUnset": "null", "laterIncrement": "int", "max": "float|int", "min": "float|int",
		"integerString": "float|int|string", "decimalString": "float|int|string",
		"topIncrement": "int", "topUnset": "null",
	})
}

func TestArrayElementIncrementDecrement(t *testing.T) {
	checkAnywhere(t, `<?php
function elementMutations(bool $condition, string $key, $unknown) {
    $a = ['x' => null, 'sibling' => 's'];
    t('prefix', ++$a['x']); t('stored', $a['x']); t('sibling', $a['sibling']);
    $copy = $a; t('copy', $copy['x']);
    $capture = fn() => $a['x']; t('capture', $capture());
    t('array', $a);
    foreach ($a as $k => $v) { t('key', $k); t('element', $v); }
    $b = ['x' => null]; t('postfix', $b['x']++); t('postfixStored', $b['x']);
    $c = ['x' => null]; t('decrement', --$c['x']); t('decrementStored', $c['x']);
    $d = ['x' => null]; if ($condition) { ++$d['x']; } t('conditional', $d['x']);
    $e = ['x' => null]; $condition && ++$e['x']; t('optional', $e['x']);
    $computed = ['x' => null, 'other' => false]; ++$computed[$key]; t('computed', $computed['x']);
    $nested = ['x' => ['y' => null], 'sibling' => 's']; ++$nested['x']['y'];
    t('nested', $nested['x']); t('nestedSibling', $nested['sibling']); t('nestedArray', $nested);
    $nestedMissing = ['x' => []]; ++$nestedMissing['x']['y']; t('nestedAbsent', $nestedMissing['other']);
    $nestedComputed = ['x' => ['y' => null]]; ++$nestedComputed[$key]['y']; t('nestedComputed', $nestedComputed['x']);
    $invalid = ['x' => []]; ++$invalid['x']; t('invalid', $invalid['x']);
    $untracked = ['x' => $unknown]; ++$untracked['x']; t('unknown', $untracked['x']);
    $reset = ['x' => null]; ++$reset['x']; $reset = ['x' => 's']; t('reset', $reset['x']);
    $loop = ['x' => null]; while ($condition) { t('loop', $loop['x']); ++$loop['x']; }
    $paren = ['x' => null]; ++($paren)['x']; t('paren', $paren['x']);
}
`, map[string]string{
		"prefix": "int", "stored": "int|null", "sibling": "string", "copy": "int|null|string",
		"capture": "int|null|string", "array": "non-empty int[]|null[]|string[]", "key": "string", "element": "int|null|string",
		"postfix": "null", "postfixStored": "int|null", "decrement": "null", "decrementStored": "null",
		"conditional": "int|null", "optional": "int|null", "computed": "?unknown",
		"nested": "?unknown", "nestedSibling": "string", "nestedArray": "non-empty array",
		"nestedAbsent": "?unknown", "nestedComputed": "?unknown",
		"invalid": "?unknown", "unknown": "?unknown", "reset": "string", "loop": "int|null", "paren": "int|null",
	})
}
