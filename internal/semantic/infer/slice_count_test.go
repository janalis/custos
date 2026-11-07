package infer_test

import "testing"

func TestSliceCardinality(t *testing.T) {
	checkShape(t, `<?php
function f($unknown,$offset,$length,$flag){
 $a=['a','b','c'];
 t('one',count(array_slice($a,1,1)));
 t('tail',count(array_slice($a,-2)));
 t('negativeLength',count(array_slice($a,0,-1)));
 t('negativeBoth',count(array_slice($a,-2,-1)));
 t('past',count(array_slice($a,8,4)));
 t('before',count(array_slice($a,-8,4)));
 t('nullLength',count(array_slice($a,1,null)));
 t('empty',count(array_slice($a,1,0)));
 t('unknownOffset',count(array_slice($a,$offset,1)));
 t('unknownLength',count(array_slice($a,1,$length)));
 t('unknownInput',count(array_slice($unknown,1,1)));
 t('range',count(array_slice($flag?['a','b']:['c'],1,1)));
 t('emptyAlt',$flag ? ['a'] : array_values([]));
 t('emptyFilter',array_filter([false,null],...[]));
}
`, map[string]string{
		"one": "int<1, 1>", "tail": "int<2, 2>", "negativeLength": "int<2, 2>", "negativeBoth": "int<1, 1>", "past": "int<0, 0>", "before": "int<3, 3>", "nullLength": "int<2, 2>", "empty": "int<0, 0>", "unknownOffset": "int<0, 3>", "unknownLength": "int<0, 3>", "unknownInput": "int", "range": "int<0, 1>", "emptyAlt": "string[]{0?: string}", "emptyFilter": "array{}",
	})
}
