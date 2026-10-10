<?php
function audit_case_0() { $s = 'abc'; $offset = -4; substr($s, $offset, 1); }
function audit_case_1() { function f(string $s, int $offset) { return substr($s, $offset, 1); } }
function audit_case_2() { $values = ['abc', 'xyz']; $i = 0; substr($values[$i++], -1, 1); }
function audit_case_3() { class C { public string $s; } function f(C $c) { return substr($c->s, -1, 1); } }
