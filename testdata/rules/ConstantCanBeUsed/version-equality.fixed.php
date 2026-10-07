<?php
$a = version_compare(PHP_VERSION, '8.1', '==');
$b = version_compare(PHP_VERSION, '8', 'eq');
$c = version_compare(PHP_VERSION, '8.1', '!=');
$d = version_compare(PHP_VERSION, '7.4', 'ne');
$e = PHP_VERSION_ID === 80100;
$f = PHP_VERSION_ID !== 70433;
$g = PHP_VERSION_ID >= 80100;
