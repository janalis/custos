<?php
$n = 0; $f = function () use (&$n) { ++$n; }; $f(); echo $n;

$n=0;$f=function()use($n){if(false){++$n;}};$f();echo $n;
