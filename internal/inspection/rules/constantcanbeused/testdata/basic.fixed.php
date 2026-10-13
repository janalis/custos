<?php
use function pi;

$sapi = PHP_SAPI;
$ver  = PHP_VERSION;
$area = 2 * M_PI * $r;

$a = PHP_VERSION_ID < 80000;
$b = PHP_VERSION_ID >= 70433;
$c = version_compare(PHP_VERSION, '8.2.0', 'eq');
$d = version_compare(\PHP_VERSION, '5.6.4', '<>');
$e = version_compare(PHP_VERSION, '7.0.123', '<=');
if (!(PHP_VERSION_ID >= 70100)) {}
