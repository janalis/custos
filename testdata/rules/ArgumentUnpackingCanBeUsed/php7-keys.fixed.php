<?php
function build(array $cols, $k) {
    $a = call_user_func_array('sprintf', $cols);
    $b = max(...[3, 9]);
    $c = max(...[0 => 3, 1 => 9]);
    $d = call_user_func_array('max', [$k => 3]);
    $e = call_user_func_array('max', ['a' => 3, 'b' => 9]);
}
