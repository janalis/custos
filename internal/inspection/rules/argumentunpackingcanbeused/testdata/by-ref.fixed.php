<?php
// Unpacking binds by-reference parameters; call_user_func_array() does not.
function sortRows(array $params, array $parts) {
    call_user_func_array('array_multisort', $params);
    return implode(...$parts);
}
