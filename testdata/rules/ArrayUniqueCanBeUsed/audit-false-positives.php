<?php
function audit_case_0() { array_keys(array_count_values(['1'])); }
function audit_case_1() { count(array_count_values(['01', '1', 1, 1.5, false])); }
function audit_case_2() { function f(array $a) { return count(array_count_values($a)); } }
