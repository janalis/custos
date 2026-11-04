package infer_test

import (
	"fmt"
	"strings"
	"testing"
)

func TestArrayElementsSurviveShapeFallback(t *testing.T) {
	var entries strings.Builder
	for i := range 34 {
		value := "1"
		if i%2 != 0 {
			value = "'text'"
		}
		fmt.Fprintf(&entries, "%s,", value)
	}
	check(t, `<?php function run(string $key, int $i, bool $choose) {
 $a = [1, 'text'];
 $b = [$key => 1, 'fixed' => 'text'];
 t('computed', $b[$key]);
 $copy = [...$b]; t('spread', $copy[$key]);
 $union = $b + [$key => false]; t('union', $union[$key]);
 $nested = [$key => [1, 2]]; t('nested', $nested[$key][0]);
 $large = [`+entries.String()+`]; t('overflow', $large[$i]);
 foreach ($large as $value) { t('foreach', $value); }
 $large[] = null; t('write', $large[$i]);
 $mixed = $choose ? $b : unknown(); t('unknown', $mixed[$key]);
 }`, map[string]string{
		"computed": "int|string", "spread": "int|string", "union": "false|int|string",
		"nested": "int", "overflow": "int|string", "foreach": "int|string",
		"write": "int|null|string", "unknown": "?unknown",
	})
}

func TestGenericArrayElementsSurviveOperations(t *testing.T) {
	check(t, `<?php
 /** @template T */
 class Box {
  /** @return T */
  public function value() {}
 }
 /** @param Box<int> $box */
 function run(Box $box, string $key) {
  $a=[$key=>$box]; t('computed',$a[$key]->value());
  $b=[...$a]; t('spread',$b[$key]->value());
  $c=$a+$b; t('union',$c[$key]->value());
  foreach ($b as $value) { t('foreach',$value->value()); }
 }`, map[string]string{"computed": "int", "spread": "int", "union": "int", "foreach": "int"})
}
