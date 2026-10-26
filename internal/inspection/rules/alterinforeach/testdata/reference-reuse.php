<?php
function alias(array $var): array
{
    foreach ($var as &$v) {
    }
    $var[] = &$v;
    foreach ($var as &$v) {
    }
    $all = [1, &$v];
    foreach ($var as &<warning descr="Unset '$v' right after the loop: it is still a reference to the last element.">$v</warning>) {
    }
    $copy = $v;
    foreach ($var as &<warning descr="Unset '$v' right after the loop: it is still a reference to the last element.">$v</warning>) {
    }
    $other = &$w;
    foreach ($var as &<warning descr="Unset '$v' right after the loop: it is still a reference to the last element.">$v</warning>) {
    }
    $cb = function () use (&$v) { $x = &$v; };
    return $var;
}
