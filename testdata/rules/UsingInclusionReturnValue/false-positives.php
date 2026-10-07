<?php
require_once __DIR__ . '/bootstrap.php';
include('helpers.php');
if ($debug) {
    require 'debug.php';
}
foreach ($plugins as $p) include $p;
@include 'optional.php';
@include_once('optional.php');
(require 'wrapped.php');
@(include 'both.php');
