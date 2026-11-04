package infer_test

import "testing"

func TestImpossibleLoopAlternativeDoesNotPoisonType(t *testing.T) {
	check(t, `<?php
 function run(bool $again) {
  $v = null;
  while ($again) {
   if ($v !== null) { t('narrowed', $v); }
   $v = 'ready';
  }
 }
 function missing($unknown, bool $again) {
  $v = $unknown;
  while ($again) {
   if ($v !== null) { t('unknown', $v); }
   $v = 'ready';
  }
 }`, map[string]string{"narrowed": "string", "unknown": "?unknown"})
}
