<?php
function sequential($o) {
    return <weak_warning descr="Replace with 'array_is_list($o-&gt;Data())'.">array_values($o->Data()) === $o->data()</weak_warning>;
}
