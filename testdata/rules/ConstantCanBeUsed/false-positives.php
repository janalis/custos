<?php
$ext  = phpversion('intl');
$host = php_uname('n');
$cls  = get_class($this);
$obj  = $this->phpversion();

$e = version_compare(PHP_VERSION, '8.2.0-RC1', '>=');
$f = version_compare(PHP_VERSION, '10.1', '>=');
$g = version_compare(PHP_VERSION, '8.2');
$h = version_compare($other, '8.2', '>=');
$i = version_compare(PHP_VERSION, '8.2', 'GE');
$j = version_compare(PHP_VERSION, '', '>=');
$k = version_compare(PHP_VERSION, '7.12', '>=');
$l = version_compare(PHP_VERSION, $min, '>=');
$m = version_compare(PHP_VERSION, '7.1', $op);

// PHP_OS sniffing needs PHP 7.2 (this fixture runs at 7.1)
$win = stripos(PHP_OS, 'win') === 0;
