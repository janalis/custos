<?php
namespace Calendar;

function strtotime($spec, $base = null) { return 0; }

$a = strtotime('now');
$b = strtotime('+1 day', \time());
$c = \Other\strtotime('now');
$d = \time();
