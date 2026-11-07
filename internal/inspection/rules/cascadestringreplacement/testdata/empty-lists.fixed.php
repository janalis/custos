<?php
function emptyLists($text) {
    $text = str_replace([':'], ['-'], $text);
    return $text;
}
function emptyReplacement($text) {
    $text = str_replace(['a', 'b', 'c'], ['', '', 'x'], $text);
    return $text;
}
function emptyLast($text) {
    $text = str_replace(['a'], ['x'], $text);
    return $text;
}
function emptySearchScalar($text) {
    $text = str_replace(['a'], ['x'], $text);
    return $text;
}
function nativeCardinality($text) {
    $search = ['a', 'b'];
    $text = str_replace([...array_values($search), 'c'], ['x', 'x', 'y'], $text);
    return $text;
}
function unusedPurePairs($text) {
    $text = str_replace(['a'], ['x'], $text);
    return $text;
}
function stringSubject(string $text, string $replacement) {
    $s = str_replace('a', $replacement, $text);
    return $s;
}
