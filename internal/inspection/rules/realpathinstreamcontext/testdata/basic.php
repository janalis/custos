<?php
$root = <warning descr="Use 'dirname(dirname(__DIR__, 1)) . '/var'' instead: realpath() fails inside stream wrappers.">realpath(dirname(__DIR__, 1) . '/../var')</warning>;
$cfg  = <warning descr="Use 'dirname(dirname($app)) . &quot;/etc&quot;' instead: realpath() fails inside stream wrappers.">realpath($app . "/../../etc")</warning>;
$odd  = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath($app . '/lib' . '/..')</warning>;
$rel  = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath('../shared')</warning>;
require <warning descr="Use ''/opt/app/boot.php'' instead: realpath() fails inside stream wrappers.">realpath('/opt/app/boot.php')</warning>;
include_once (<warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath($entry)</warning>);
$deep = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath(trim($x . '/../y'))</warning>;
$ok   = realpath($app . '/cache');
$two  = realpath('/tmp', 'x');
