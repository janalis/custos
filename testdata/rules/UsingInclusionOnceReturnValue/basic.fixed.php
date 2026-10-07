<?php
$routes = require __DIR__ . '/routes.php';
$ok = include $first ||
    include $second;
while (include ($plugin)) {
    register((require $plugin));
}
require_once __DIR__ . '/bootstrap.php';
include_once('helpers.php');
$cfg = (require 'config.php');
