<?php
// Unpacking binds by-reference parameters; call_user_func_array() does not.
function sortRows(array $params, array $parts) {
    <warning descr="Call 'array_multisort(...$params)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('array_multisort', $params)</warning>;
    return <warning descr="Call 'implode(...$parts)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('implode', $parts)</warning>;
}
