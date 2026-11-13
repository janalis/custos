<?php
// Comparisons the owning rules do not report stay this rule's: the
// function form of collator_compare(), a collation result through a local
// or compared with !==, and array_search() in parentheses.
function gaps($c, Collator $col, string $who, array $roles): array
{
    $r = $col->compare('a', 'b');
    return [
        <error descr="This comparison is always false: false|int and true share no value.">collator_compare($c, 'a', 'b') === true</error>,
        <error descr="This comparison is always false: false|int and true share no value.">$r === true</error>,
        <error descr="This comparison is always true: false|int and true share no value.">$col->compare('a', 'b') !== true</error>,
        <error descr="This comparison is always false: false|int|string and true share no value.">(array_search($who, $roles, true)) === true</error>,
    ];
}
