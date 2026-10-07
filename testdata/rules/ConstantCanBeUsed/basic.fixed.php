<?php
use function pi;

$sapi = PHP_SAPI;
$ver  = PHP_VERSION;
$area = 2 * M_PI * $r;

$a = PHP_VERSION_ID < 80000;
$b = PHP_VERSION_ID >= 70433;
$c = PHP_VERSION_ID === 80200;
$d = PHP_VERSION_ID !== 50604;
$e = PHP_VERSION_ID <= 700123;
if (!(PHP_VERSION_ID >= 70100)) {}
