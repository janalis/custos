<?php
// Still definite: nothing can rewrite the variable behind the assignments.
function plainLocal(string $line) {
    $glue = '';
    return <error descr="Supply a nonempty separator.">explode($glue, $line)</error>;
}
function copied(string $line, string $query) {
    $empty = '';
    $glue = $empty;
    $glue = $glue;
    parse_str($query, $fields);
    $count = 0;
    $count++;
    $other = '';
    $unrelated = function () use (&$other, $count) { $other = ','; };
    $nested = function () { $glue = ','; $glue++; };
    $pairs = ['a' => $count, 'b' => &$other];
    return <error descr="Supply a nonempty separator.">explode($glue, $line)</error>;
}
