<?php
function swap(array $from, array $to, $text) {
    $text = str_replace('<', '&lt;', $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace($from, $to, $text)</warning>;
    return $text;
}

function sheetName(array $bad, $name) {
    $name = str_replace($bad, ' ', $name);
    return <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('\'', '', $name)</warning>;
}

function strip(array $bad, $name) {
    $name = str_replace($bad, '', $name);
    return <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('\'', '', $name)</warning>;
}

function nested(array $bad, $name) {
    return str_replace('\'', '', <warning descr="Fold this nested str_replace() into the enclosing call.">str_replace($bad, ' ', $name)</warning>);
}
