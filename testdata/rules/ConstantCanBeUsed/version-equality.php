<?php
$a = version_compare(PHP_VERSION, '8.1', '==');
$b = version_compare(PHP_VERSION, '8', 'eq');
$c = version_compare(PHP_VERSION, '8.1', '!=');
$d = version_compare(PHP_VERSION, '7.4', 'ne');
$e = <weak_warning descr="Replace with 'PHP_VERSION_ID === 80100'.">version_compare(PHP_VERSION, '8.1.0', '==')</weak_warning>;
$f = <weak_warning descr="Replace with 'PHP_VERSION_ID !== 70433'.">version_compare(PHP_VERSION, '7.4.33', '!=')</weak_warning>;
$g = <weak_warning descr="Replace with 'PHP_VERSION_ID >= 80100'.">version_compare(PHP_VERSION, '8.1', '>=')</weak_warning>;
