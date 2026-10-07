<?php
$a = version_compare(PHP_VERSION, '8.1', '==');
$b = version_compare(PHP_VERSION, '8', 'eq');
$c = version_compare(PHP_VERSION, '8.1', '!=');
$d = version_compare(PHP_VERSION, '7.4', 'ne');
$e = PHP_VERSION_ID === 80100;
$f = PHP_VERSION_ID !== 70433;
$g = PHP_VERSION_ID >= 80100;
$h = !(PHP_VERSION_ID < 80000);
$i = (PHP_VERSION_ID >= 80000) . 'x';
$j = PHP_VERSION_ID >= 80000 && $h;
$k = (bool) (PHP_VERSION_ID >= 80000);
// not reported: not a M[.m[.p]] version with single-digit major/minor
$l = version_compare(PHP_VERSION, '8.1.2.3', '>=');
$m = version_compare(PHP_VERSION, '10.0', '>=');
$n = version_compare(PHP_VERSION, '8.12', '>=');
$o = version_compare(PHP_VERSION, '8..1', '>=');
$p = version_compare(PHP_VERSION, '8.x', '>=');
