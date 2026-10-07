<?php
use function pi;

$sapi = <weak_warning descr="Use the PHP_SAPI constant instead of this call.">php_sapi_name()</weak_warning>;
$ver  = <weak_warning descr="Use the PHP_VERSION constant instead of this call.">\phpversion()</weak_warning>;
$area = 2 * <weak_warning descr="Use the M_PI constant instead of this call.">pi()</weak_warning> * $r;

$a = <weak_warning descr="Replace with 'PHP_VERSION_ID < 80000'.">version_compare(PHP_VERSION, '8', 'lt')</weak_warning>;
$b = <weak_warning descr="Replace with 'PHP_VERSION_ID >= 70433'.">version_compare(PHP_VERSION, "7.4.33", 'ge')</weak_warning>;
$c = <weak_warning descr="Replace with 'PHP_VERSION_ID === 80200'.">version_compare(PHP_VERSION, '8.2.0', 'eq')</weak_warning>;
$d = <weak_warning descr="Replace with 'PHP_VERSION_ID !== 50604'.">version_compare(\PHP_VERSION, '5.6.4', '<>')</weak_warning>;
$e = <weak_warning descr="Replace with 'PHP_VERSION_ID <= 700123'.">version_compare(PHP_VERSION, '7.0.123', '<=')</weak_warning>;
if (!<weak_warning descr="Replace with 'PHP_VERSION_ID >= 70100'.">version_compare(PHP_VERSION, '7.1', '>')</weak_warning>) {}
