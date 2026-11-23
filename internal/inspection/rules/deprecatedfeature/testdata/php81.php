<?php
// At a PHP 8.1 target only what 8.1 and earlier deprecate is reported.
$day = <warning descr="Replace strftime(), deprecated as of PHP 8.1.">strftime</warning>('%A');
$rows = <warning descr="Replace pg_numrows(), deprecated as of PHP 8.0.">pg_numrows</warning>($result);
$text = utf8_encode($raw);
$r = lcg_value();
function open(ReflectionProperty $p): void { $p->setAccessible(true); }
