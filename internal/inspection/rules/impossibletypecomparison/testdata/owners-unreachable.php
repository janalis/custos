<?php
// The owning rules skip unreachable code: the comparisons stay this rule's
// there, and theirs where the code runs, closures included.
function unreachable($parser, string $data, Collator $c): array
{
    $run = fn() => [
        <warning descr="Compare XML parsing failure with integer zero.">xml_parse($parser, $data) === false</warning>,
        <warning descr="Compare collation results explicitly.">$c->compare('a', 'b') === true</warning>,
    ];
    return $run();
    return [
        <error descr="This comparison is always false: int and false share no value.">xml_parse($parser, $data) === false</error>,
        <error descr="This comparison is always false: false|int and true share no value.">$c->compare('a', 'b') === true</error>,
    ];
}
