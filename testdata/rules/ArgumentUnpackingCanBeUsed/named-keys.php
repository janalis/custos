<?php
function build(array $opts) {
    $a = <warning descr="Call 'greet(...$opts)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('greet', $opts)</warning>;
    $b = <warning descr="Call 'greet(...['name' => 'Ann'])' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('greet', ['name' => 'Ann'])</warning>;
}
