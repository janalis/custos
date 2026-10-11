<?php
$a = PHP_VERSION_ID >= 70100;
$b = PHP_VERSION_ID < 70100;
$c = PHP_VERSION_ID >= 80000;
$d = PHP_VERSION_ID < 80000;
$e = PHP_VERSION_ID < 70400;
$f = version_compare(PHP_VERSION, '7.1.0', '>');
$g = version_compare(PHP_VERSION, '7.4.33', '<=');
// a full version keeps its fix only for `<` and `>=`
$h = PHP_VERSION_ID < 80102;
$i = PHP_VERSION_ID >= 80102;
$j = version_compare(PHP_VERSION, '8.1.2', 'gt');
$k = !version_compare(PHP_VERSION, '8.1.2', 'le');
