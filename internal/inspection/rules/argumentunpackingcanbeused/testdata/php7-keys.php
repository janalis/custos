<?php
function build(array $cols, $k) {
    $a = <warning descr="Call 'sprintf(...$cols)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('sprintf', $cols)</warning>;
    $b = <warning descr="Call 'max(...[3, 9])' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('max', [3, 9])</warning>;
    $c = <warning descr="Call 'max(...[0 => 3, 1 => 9])' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('max', [0 => 3, 1 => 9])</warning>;
    $d = <warning descr="Call 'max(...[$k => 3])' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('max', [$k => 3])</warning>;
    $e = call_user_func_array('max', ['a' => 3, 'b' => 9]);
}
