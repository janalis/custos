<?php
$a = <weak_warning descr="Replace with 'PHP_VERSION_ID >= 70100'.">version_compare(PHP_VERSION, '7.1', '>')</weak_warning>;
$b = <weak_warning descr="Replace with 'PHP_VERSION_ID < 70100'.">version_compare(PHP_VERSION, '7.1', '<=')</weak_warning>;
$c = <weak_warning descr="Replace with 'PHP_VERSION_ID >= 80000'.">version_compare(PHP_VERSION, '8', 'gt')</weak_warning>;
$d = <weak_warning descr="Replace with 'PHP_VERSION_ID < 80000'.">version_compare(PHP_VERSION, '8', 'le')</weak_warning>;
$e = <weak_warning descr="Replace with 'PHP_VERSION_ID < 70400'.">version_compare(PHP_VERSION, '7.4', '<')</weak_warning>;
$f = <weak_warning descr="Compare PHP_VERSION_ID instead: 'PHP_VERSION_ID > 70100' ignores build suffixes.">version_compare(PHP_VERSION, '7.1.0', '>')</weak_warning>;
$g = <weak_warning descr="Compare PHP_VERSION_ID instead: 'PHP_VERSION_ID <= 70433' ignores build suffixes.">version_compare(PHP_VERSION, '7.4.33', '<=')</weak_warning>;
// a full version keeps its fix only for `<` and `>=`
$h = <weak_warning descr="Replace with 'PHP_VERSION_ID < 80102'.">version_compare(PHP_VERSION, '8.1.2', 'lt')</weak_warning>;
$i = <weak_warning descr="Replace with 'PHP_VERSION_ID >= 80102'.">version_compare(PHP_VERSION, '8.1.2', '>=')</weak_warning>;
$j = <weak_warning descr="Compare PHP_VERSION_ID instead: 'PHP_VERSION_ID > 80102' ignores build suffixes.">version_compare(PHP_VERSION, '8.1.2', 'gt')</weak_warning>;
$k = !<weak_warning descr="Compare PHP_VERSION_ID instead: 'PHP_VERSION_ID <= 80102' ignores build suffixes.">version_compare(PHP_VERSION, '8.1.2', 'le')</weak_warning>;
