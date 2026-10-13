<?php
function describe($value): string {
    // gettype(1) is 'integer' but get_debug_type(1) is 'int': no automatic rewrite.
    $t = <weak_warning descr="Use 'get_debug_type($value)' instead (scalar type names differ from gettype()).">is_object($value) ? get_class($value) : gettype($value)</weak_warning>;
    return $t === 'integer' ? 'number' : $t;
}
