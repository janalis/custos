<?php
// Reported without a fix: the merged call could not keep the result.

// $search may be a string or an array: neither one element nor a spread.
function untyped($text, $search, $replace) {
    $text = str_replace($search, $replace, $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace("\n", ' ', $text)</warning>;
    return $text;
}

function either(string $text, array|string $bad) {
    return str_replace('x', 'y', <warning descr="Fold this nested str_replace() into the enclosing call.">str_replace($bad, '-', $text)</warning>);
}

// 'b' is paired with '' and '2' is ignored; merged lists would shift.
function shortList($s) {
    $s = str_replace(['a', 'b'], ['1'], $s);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace('c', '3', $s)</warning>;
    return $s;
}

function longList($s) {
    return str_replace('c', '3', <warning descr="Fold this nested str_replace() into the enclosing call.">str_replace(['a'], ['1', '2'], $s)</warning>);
}

// The merged call would read the second key before taking the template.
function order(ArrayIterator $q) {
    $s = str_replace('a', 'b', take($q));
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(take($q), 'x', $s)</warning>;
    return $s;
}

function take(ArrayIterator $q): string {
    $v = $q->current();
    $q->next();
    return $v;
}

// Still fixed: typed strings, a short list ending the merge, one pure call.
function typed(string $text, string $from, string $to) {
    $text = str_replace($from, $to, $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['a', 'b'], ['1'], $text)</warning>;
    return $text;
}

function pure(string $text) {
    $text = str_replace('a', 'b', $text);
    return <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(chr(10), ' ', $text)</warning>;
}
