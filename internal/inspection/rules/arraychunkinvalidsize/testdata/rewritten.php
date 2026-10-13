<?php
function plainZero(array $items) {
    $size = 0;
    return <error descr="Supply a positive chunk size.">array_chunk($items, $size)</error>;
}
function closureByValue(array $items) {
    $size = 0;
    $read = function () use ($size) { return $size; };
    return <error descr="Supply a positive chunk size.">array_chunk($items, $size)</error>;
}
function otherCounter(array $items, array $headers) {
    $size = 0;
    $seen = 0;
    foreach ($headers as $h) {
        $seen++;
    }
    return <error descr="Supply a positive chunk size.">array_chunk($items, $size)</error>;
}
function countedInLoop(array $items, array $headers) {
    $size = 0;
    foreach ($headers as $h) {
        $size++;
    }
    return array_chunk($items, $size);
}
function preIncrement(array $items) {
    $size = 0;
    ++$size;
    return array_chunk($items, ($size));
}
function decremented(array $items) {
    $size = -3;
    $size--;
    return array_chunk($items, $size);
}
function closureByReference(array $items) {
    $size = 0;
    $set = function (int $n) use (&$size) { $size = $n; };
    $set(2);
    return array_chunk($items, $size);
}
function extracted(array $items, array $options) {
    $size = 0;
    EXTRACT($options);
    return array_chunk($items, $size);
}
function referencedItem(array $items) {
    $size = 0;
    foreach ([&$size] as &$slot) {
        $slot = 3;
    }
    return array_chunk($items, $size);
}
function copiedLiteral(array $items) {
    $zero = 0;
    $size = $zero;
    return <error descr="Supply a positive chunk size.">array_chunk($items, $size)</error>;
}
function copiedCounter(array $items, array $headers) {
    $seen = 0;
    foreach ($headers as $h) {
        $seen++;
    }
    $size = $seen;
    return array_chunk($items, $size);
}
