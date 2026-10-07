<?php
$a = <weak_warning descr="Replace with 'PHP_VERSION_ID >= 70100'.">version_compare(PHP_VERSION, '7.1', '>')</weak_warning>;
$b = <weak_warning descr="Replace with 'PHP_VERSION_ID < 70100'.">version_compare(PHP_VERSION, '7.1', '<=')</weak_warning>;
$c = <weak_warning descr="Replace with 'PHP_VERSION_ID >= 80000'.">version_compare(PHP_VERSION, '8', 'gt')</weak_warning>;
$d = <weak_warning descr="Replace with 'PHP_VERSION_ID < 80000'.">version_compare(PHP_VERSION, '8', 'le')</weak_warning>;
$e = <weak_warning descr="Replace with 'PHP_VERSION_ID < 70400'.">version_compare(PHP_VERSION, '7.4', '<')</weak_warning>;
$f = <weak_warning descr="Replace with 'PHP_VERSION_ID > 70100'.">version_compare(PHP_VERSION, '7.1.0', '>')</weak_warning>;
$g = <weak_warning descr="Replace with 'PHP_VERSION_ID <= 70433'.">version_compare(PHP_VERSION, '7.4.33', '<=')</weak_warning>;
