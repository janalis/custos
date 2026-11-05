package infer_test

import "testing"

func TestPackedBuiltinReturnRoles(t *testing.T) {
	checkShape(t, `<?php
function f($unknown) {
 t('clock',hrtime(...['as_number'=>true]));
 t('time',gettimeofday(...[true]));
 t('printed',print_r(...['value'=>'x','return'=>true]));
 t('normalPrint',print_r(...['x']));
 t('replaced',str_replace(...['subject'=>['a'],'replace'=>'b','search'=>'a']));
 t('path',pathinfo(...['path'=>'/a/b']));
 t('values',array_values(...[['a'=>1]]));
 t('keys',array_rand(...[['a'=>'x']]));
 t('invalid',hrtime(...['wrong'=>true]));
 t('unknown',hrtime(...$unknown));
}
`, map[string]string{"clock": "false|float|int", "time": "float", "printed": "string", "normalPrint": "true", "replaced": "array", "path": "array{dirname?: string, basename: string, extension?: string, filename: string}", "values": "non-empty int[]", "keys": "int|string", "invalid": "false|float|int|int[]", "unknown": "false|float|int|int[]"})
}

func TestTransformedArrayCount(t *testing.T) {
	checkShape(t, `<?php
function f($c, array $unknown, $mode) {
 $xs=['a'=>1,'b'=>2];
 t('values',count(array_values($xs)));
 t('reverse',count(array_reverse($xs)));
 t('unique',count(array_unique($xs)));
 t('filter',count(array_filter($xs)));
 t('slice',count(array_slice($xs,1)));
 t('zero',sizeof(array_values([])));
 t('branch',count($c?['a'=>1]:['b'=>2]));
 t('emptyBranch',count($c?array_values($xs):[]));
 t('normal',count(array_values($xs), COUNT_NORMAL));
 t('normalZero',count(array_values($xs),0));
 t('recursive',count(array_values($xs), COUNT_RECURSIVE));
 t('unknown',count(array_values($unknown)));
 t('mode',count(array_values($xs),$mode));
 $ys=array_values($xs);$ys[]=3;t('mutation',count($ys));
}
`, map[string]string{"values": "int<2, 2>", "reverse": "int<2, 2>", "unique": "int<1, 2>", "filter": "int<0, 2>", "slice": "int<0, 2>", "zero": "int<0, 0>", "branch": "int<1, 1>", "emptyBranch": "int<0, 2>", "normal": "int<2, 2>", "normalZero": "int<2, 2>", "recursive": "int", "unknown": "int", "mode": "int", "mutation": "int"})
}

func TestCountModeShadowing(t *testing.T) {
	checkShape(t, `<?php
namespace App;
const COUNT_NORMAL = 1;
t('shadow',count([['x']],COUNT_NORMAL));
t('global',count([['x']],\COUNT_NORMAL));
`, map[string]string{"shadow": "int", "global": "int<1, 1>"})
}
