<?php
function escape(array $a, array $b, $text) {
    $text = str_replace($a, $b, $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(array_keys($b), $a, $text)</warning>;
    return $text;
}

function twice(array $from1, array $to1, array $from2, array $to2, $text) {
    $text = str_replace($from1, $to1, $text);
    return <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace($from2, $to2, $text)</warning>;
}
