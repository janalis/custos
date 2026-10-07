<?php

function describe(mixed $value): string
{
    return <weak_warning descr="Use 'get_debug_type($value)' instead (scalar type names differ from gettype()).">Is_Object($value) ? GET_CLASS($value) : GetType($value)</weak_warning>;
}
