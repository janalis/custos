<?php
function inspect(array $cfg, array $other) {
    return [
        <weak_warning descr="Replace with 'array_is_list($cfg)'.">array_values($cfg) === $cfg</weak_warning>,
        <weak_warning descr="Replace with '!array_is_list($cfg)'.">$cfg !== array_values($cfg)</weak_warning>,
        <weak_warning descr="Replace with '!\array_is_list($cfg)'.">\array_values($cfg) !== $cfg</weak_warning>,
        <weak_warning descr="Replace with '$cfg !== [] && array_is_list($cfg)'.">array_keys($cfg) === range(0, count($cfg) - 1)</weak_warning>,
        <weak_warning descr="Replace with '$cfg === [] || !array_is_list($cfg)'.">range(0, count($cfg) - 1) !== array_keys($cfg)</weak_warning>,
        <weak_warning descr="Replace with 'array_is_list($this->rows['a'])'.">array_values($this->rows['a']) === $this->rows[ 'a' ]</weak_warning>,
    ];
}

function wrapped(array $cfg): bool
{
    return $cfg && !(<weak_warning descr="Replace with '$cfg !== [] && array_is_list($cfg)'.">array_keys($cfg) === range(0, count($cfg) - 1)</weak_warning>)
        || <weak_warning descr="Replace with '$cfg === [] || !array_is_list($cfg)'.">array_keys($cfg) !== range(0, count($cfg) - 1)</weak_warning> === false;
}
