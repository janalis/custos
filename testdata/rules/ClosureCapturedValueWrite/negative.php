<?php
$n = 0; $f = function () use (&$n) { ++$n; }; $f(); echo $n;

$n=0;$f=function()use($n){if(false){++$n;}};$f();echo $n;

function countdown(int $left): void {
    $tick = function () use ($left) {
        while ($left-- > 0) {
            echo '.';
        }
    };
    $tick();
    echo $left;
}

function badge(string $who): void {
    $print = function () use ($who) {
        $who = ucfirst(ltrim($who));
        echo "<$who>";
    };
    $print();
    echo $who;
}
