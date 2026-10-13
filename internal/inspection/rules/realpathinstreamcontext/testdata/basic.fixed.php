<?php
$root = dirname(dirname(__DIR__, 1)) . '/var';
$cfg  = realpath($app . "/../../etc");
$odd  = realpath($app . '/lib' . '/..');
$rel  = realpath('../shared');
require '/opt/app/boot.php';
include_once (realpath($entry));
$deep = realpath(trim($x . '/../y'));
$ok   = realpath($app . '/cache');
$two  = realpath('/tmp', 'x');
