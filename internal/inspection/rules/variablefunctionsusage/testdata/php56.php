<?php
function run($hook, $name) {
    // Before 7.0, $hook() cannot call a 'Cls::m' string.
    call_user_func($hook, 1);
    $byName = 'Logger::write';
    call_user_func($byName, 2);
    call_user_func($name ? 'trim' : 'rtrim', 3);
    $closure = function ($x) { return $x; };
    <weak_warning descr="Call it directly: '$closure(4)'.">call_user_func($closure, 4)</weak_warning>;
    $pair = array(new Logger(), 'write');
    <weak_warning descr="Call it directly: '$pair(5)'.">call_user_func($pair, 5)</weak_warning>;
}
