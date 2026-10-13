<?php
require_once 'a.php';
include_once 'b.php';
$x = require 'c.php';
$y = include 'd.php';
// Silenced statements discard the result too.
@include_once 'lang/' . $iso . '.php';
(@require_once 'e.php');
