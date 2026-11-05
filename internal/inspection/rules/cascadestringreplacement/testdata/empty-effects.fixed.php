<?php
function keepEvaluation($text, $replacement) {
    $text = str_replace([], computeReplacement(), $text);
    $text = str_replace('a', 'b', $text);
    return $text;
}
function keepAssignment($text) {
    $text = str_replace(['a'], $replacement = [], $text);
    $text = str_replace('b', 'c', $text);
    return [$text, $replacement];
}
function keepArrayEvaluation($text) {
    $text = str_replace([], [computeReplacement()], $text);
    $text = str_replace('a', 'b', $text);
    return $text;
}
function keepCoercion($text) {
    $text = str_replace([], null, $text);
    $text = str_replace('a', 'b', $text);
    return $text;
}
function keepKeyDeprecation($text) {
    $text = str_replace([], [1.5 => 'x'], $text);
    $text = str_replace('a', 'b', $text);
    return $text;
}
function keepSubjectConversion($text, string $replacement) {
    $s = str_replace([], [], $text);
    $s = str_replace('a', $replacement, $s);
    return $s;
}
function keepElementConversions(string $text, array $search) {
    $s = str_replace($search, [], $text);
    $s = str_replace('a', 'x', $s);
    return $s;
}
function keepScalarValidation(string $text) {
    $s = str_replace([], [], $text);
    $s = str_replace('a', false, $s);
    return $s;
}
function keepUnknownElementTypes(string $text, array $search) {
    $s = str_replace([], [], $text);
    $s = str_replace($search, ['x'], $s);
    return $s;
}
function keepNestedArrayEvaluation($text) {
    $s = str_replace([], [[1.5 => 'x']], $text);
    $s = str_replace('a', 'b', $s);
    return $s;
}
function keepCastedSubject($object, string $replacement) {
    $s = str_replace([], [], (string)$object);
    $s = str_replace('a', $replacement, $s);
    return $s;
}
function keepCastedReplacement(string $text, $object) {
    $s = str_replace([], [], $text);
    $s = str_replace('a', (string)$object, $s);
    return $s;
}
