<?php
namespace Clock;

function time() { return 0; }

$a = \time();
$b = \time();
$c = \strtotime('+1 day');
