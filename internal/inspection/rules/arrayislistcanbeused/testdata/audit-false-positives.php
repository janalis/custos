<?php
function audit_case_0() { function nextArray(): array { static $i = 0; return [++$i]; } array_values(nextArray()) === nextArray(); }
function audit_case_1() { array_values($a[$i++]) === $a[$i++]; }
function audit_case_2() { array_values($$name) === $$name; }
function audit_case_3() { array_values($obj->items) === $obj->items; }
