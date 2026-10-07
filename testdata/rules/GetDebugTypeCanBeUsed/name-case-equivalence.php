<?php
function typeName($o) {
    return <weak_warning descr="Use 'get_debug_type($o-&gt;Value())' instead (scalar type names differ from gettype()).">is_object($o->Value()) ? get_class($o->value()) : gettype($o->VALUE())</weak_warning>;
}
