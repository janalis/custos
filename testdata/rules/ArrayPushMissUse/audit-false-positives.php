<?php
function audit_case_0() { $a = [2 => 'old']; $a[count($a)] = 'new'; }
function audit_case_1() { $a = ['old', 'last']; unset($a[1]); $a[count($a)] = 'new'; }
function audit_case_2() { function f(array $a) { $a[count($a)] = 'new'; } }
function audit_case_3() { $a = []; $v = ($a[count($a)] = 'new'); }
function audit_case_4() { $a = []; $a[count($a)] = mutate($a); }
function audit_case_5() { $a->x[count($a->x)] = 'new'; }
function audit_case_6() { $a = &$b; $a[count($a)] = 'new'; }
function audit_case_7() { $a = [...$b]; $a[count($a)] = 'new'; }
function audit_case_8() { $a = [&$b]; $a[count($a)] = 'new'; }
function audit_case_9() { ping(); $a[count($a)] = 'new'; }
function audit_case_10() { echo 'ok'; $a[count($a)] = 'new'; }
function audit_case_11() { $b = []; $a[count($a)] = 'new'; }
function audit_case_12() { $a += []; $a[count($a)] = 'new'; }
function audit_case_13() { $a = source(); $a[count($a)] = 'new'; }
function audit_case_14() { $$a[count($$a)] = 'new'; }
