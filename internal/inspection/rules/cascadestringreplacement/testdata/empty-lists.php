<?php
function emptyLists($text) {
    $text = str_replace([], [], $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace([':'], ['-'], $text)</warning>;
    return $text;
}
function emptyReplacement($text) {
    $text = str_replace(['a', 'b'], [], $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['c'], ['x'], $text)</warning>;
    return $text;
}
function emptyLast($text) {
    $text = str_replace(['a'], ['x'], $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace([], [], $text)</warning>;
    return $text;
}
function emptySearchScalar($text) {
    $text = str_replace([], '-', $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['a'], ['x'], $text)</warning>;
    return $text;
}
function nativeCardinality($text) {
    $search = ['a', 'b'];
    $text = str_replace($search, 'x', $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['c'], ['y'], $text)</warning>;
    return $text;
}
function unusedPurePairs($text) {
    $text = str_replace([], ['unused' => null, 'other' => true], $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['a'], ['x'], $text)</warning>;
    return $text;
}
function stringSubject(string $text, string $replacement) {
    $s = str_replace([], [], $text);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('a', $replacement, $s)</warning>;
    return $s;
}
