<?php
// Reported without a fix: the merged call could not keep the result.

// $search may be a string or an array: neither one element nor a spread.
function untyped($text, $search, $replace) {
    $text = str_replace($search, $replace, $text);
    $text = str_replace("\n", ' ', $text);
    return $text;
}

function either(string $text, array|string $bad) {
    return str_replace('x', 'y', str_replace($bad, '-', $text));
}

// 'b' is paired with '' and '2' is ignored; merged lists would shift.
function shortList($s) {
    $s = str_replace(['a', 'b'], ['1'], $s);
    $s = str_replace('c', '3', $s);
    return $s;
}

function longList($s) {
    return str_replace('c', '3', str_replace(['a'], ['1', '2'], $s));
}

// The merged call would read the second key before taking the template.
function order(ArrayIterator $q) {
    $s = str_replace('a', 'b', take($q));
    $s = str_replace(take($q), 'x', $s);
    return $s;
}

function take(ArrayIterator $q): string {
    $v = $q->current();
    $q->next();
    return $v;
}

// Still fixed: typed strings, a short list ending the merge, one pure call.
function typed(string $text, string $from, string $to) {
    $text = str_replace([$from, 'a', 'b'], [$to, '1'], $text);
    return $text;
}

function pure(string $text) {
    return str_replace(['a', chr(10)], ['b', ' '], $text);
}
