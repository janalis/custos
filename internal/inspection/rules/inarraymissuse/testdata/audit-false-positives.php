<?php
function audit_case_0() { $strict = true; in_array('1', [1], $strict); }
function audit_case_1() { in_array('1', [1], mode()); }
function audit_case_2() { in_array(0, array_keys($a), $strict); }
function audit_case_3() { in_array('1', [key() => 1], true); }
