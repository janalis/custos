<?php
function run($hook, $name) {
    // Before 7.0, $hook() cannot call a 'Cls::m' string.
    call_user_func($hook, 1);
    $byName = 'Logger::write';
    call_user_func($byName, 2);
    call_user_func($name ? 'trim' : 'rtrim', 3);
    $closure = function ($x) { return $x; };
    $closure(4);
    $pair = array(new Logger(), 'write');
    $pair(5);
}
