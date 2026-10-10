<?php
use function pcntl_fork as builtinCall10;
$pid=builtinCall10(); if (<warning descr="Terminate or separate the fork child branch.">$pid===0</warning>) { echo "child"; } echo "continuation";
