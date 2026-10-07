<?php
function describe($value): string {
    // gettype(1) is 'integer' but get_debug_type(1) is 'int': no automatic rewrite.
    $t = is_object($value) ? get_class($value) : gettype($value);
    return $t === 'integer' ? 'number' : $t;
}
