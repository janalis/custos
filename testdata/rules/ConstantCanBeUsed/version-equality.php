<?php
$a = version_compare(PHP_VERSION, '8.1', '==');
$b = version_compare(PHP_VERSION, '8', 'eq');
$c = version_compare(PHP_VERSION, '8.1', '!=');
$d = version_compare(PHP_VERSION, '7.4', 'ne');
$e = <weak_warning descr="Replace with 'PHP_VERSION_ID === 80100'.">version_compare(PHP_VERSION, '8.1.0', '==')</weak_warning>;
$f = <weak_warning descr="Replace with 'PHP_VERSION_ID !== 70433'.">version_compare(PHP_VERSION, '7.4.33', '!=')</weak_warning>;
$g = <weak_warning descr="Replace with 'PHP_VERSION_ID >= 80100'.">version_compare(PHP_VERSION, '8.1', '>=')</weak_warning>;
$h = !<weak_warning descr="Replace with 'PHP_VERSION_ID < 80000'.">version_compare(PHP_VERSION, '8.0.0', '<')</weak_warning>;
$i = <weak_warning descr="Replace with 'PHP_VERSION_ID >= 80000'.">version_compare(PHP_VERSION, '8.0.0', '>=')</weak_warning> . 'x';
$j = <weak_warning descr="Replace with 'PHP_VERSION_ID >= 80000'.">version_compare(PHP_VERSION, '8.0.0', '>=')</weak_warning> && $h;
$k = (bool) <weak_warning descr="Replace with 'PHP_VERSION_ID >= 80000'.">version_compare(PHP_VERSION, '8.0.0', '>=')</weak_warning>;
// not reported: not a M[.m[.p]] version with single-digit major/minor
$l = version_compare(PHP_VERSION, '8.1.2.3', '>=');
$m = version_compare(PHP_VERSION, '10.0', '>=');
$n = version_compare(PHP_VERSION, '8.12', '>=');
$o = version_compare(PHP_VERSION, '8..1', '>=');
$p = version_compare(PHP_VERSION, '8.x', '>=');
