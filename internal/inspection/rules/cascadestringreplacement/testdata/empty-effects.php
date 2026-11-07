<?php
function keepEvaluation($text, $replacement) {
    $text = str_replace([], computeReplacement(), $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('a', 'b', $text)</warning>;
    return $text;
}
function keepAssignment($text) {
    $text = str_replace(['a'], $replacement = [], $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('b', 'c', $text)</warning>;
    return [$text, $replacement];
}
function keepArrayEvaluation($text) {
    $text = str_replace([], [computeReplacement()], $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('a', 'b', $text)</warning>;
    return $text;
}
function keepCoercion($text) {
    $text = str_replace([], null, $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('a', 'b', $text)</warning>;
    return $text;
}
function keepKeyDeprecation($text) {
    $text = str_replace([], [1.5 => 'x'], $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('a', 'b', $text)</warning>;
    return $text;
}
function keepSubjectConversion($text, string $replacement) {
    $s = str_replace([], [], $text);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('a', $replacement, $s)</warning>;
    return $s;
}
function keepElementConversions(string $text, array $search) {
    $s = str_replace($search, [], $text);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('a', 'x', $s)</warning>;
    return $s;
}
function keepScalarValidation(string $text) {
    $s = str_replace([], [], $text);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('a', false, $s)</warning>;
    return $s;
}
function keepUnknownElementTypes(string $text, array $search) {
    $s = str_replace([], [], $text);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace($search, ['x'], $s)</warning>;
    return $s;
}
function keepNestedArrayEvaluation($text) {
    $s = str_replace([], [[1.5 => 'x']], $text);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('a', 'b', $s)</warning>;
    return $s;
}
function keepCastedSubject($object, string $replacement) {
    $s = str_replace([], [], (string)$object);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('a', $replacement, $s)</warning>;
    return $s;
}
function keepCastedReplacement(string $text, $object) {
    $s = str_replace([], [], $text);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('a', (string)$object, $s)</warning>;
    return $s;
}
