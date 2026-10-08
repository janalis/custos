<?php
$root = dirname($dir);
$up = dirname(dirname(__DIR__));
// `..cache` is a directory name, not the parent.
$cache = realpath(__DIR__ . '/..cache');
