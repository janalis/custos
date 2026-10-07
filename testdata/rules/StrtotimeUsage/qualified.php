<?php
namespace Clock;

function time() { return 0; }

$a = <warning descr="Call time() instead of parsing 'now'.">\strtotime('now')</warning>;
$b = <warning descr="Call time() instead of parsing 'now'.">strtotime('now')</warning>;
$c = <warning descr="The base timestamp already defaults to the current time; drop the time() argument.">\strtotime('+1 day', \time())</warning>;
