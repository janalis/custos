<?php
function swap(array $from, array $to, $text) {
    $text = str_replace('<', '&lt;', $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace($from, $to, $text)</warning>;
    return $text;
}
