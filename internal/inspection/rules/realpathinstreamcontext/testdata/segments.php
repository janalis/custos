<?php
$root = <warning descr="Use 'dirname($dir)' instead: realpath() fails inside stream wrappers.">realpath($dir . '/..')</warning>;
$up = <warning descr="Use 'dirname(dirname(__DIR__))' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../..')</warning>;
// `..cache` is a directory name, not the parent.
$cache = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath(__DIR__ . '/..cache')</warning>;
