<?php
use function pcntl_fork as builtinCall10;
$pid=builtinCall10(); if (<warning descr="Handle fork failure separately.">$pid</warning>) { echo "parent"; } else { echo "child"; }
