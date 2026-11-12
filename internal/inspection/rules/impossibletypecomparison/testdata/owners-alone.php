<?php
// The same comparisons when the rules that own them do not run.
function owners(array $roles, string $who, $parser, string $data, Collator $c): array
{
    $found = array_search($who, $roles, true);
    return [
        <error descr="This comparison is always false: false|int|string and true share no value.">array_search($who, $roles, true) === true</error>,
        <error descr="This comparison is always true: true and false|int|string share no value.">true !== array_search($who, $roles, true)</error>,
        <error descr="This comparison is always true: false|int|string and true share no value.">$found !== true</error>,
        <error descr="This comparison is always false: int and false share no value.">xml_parse($parser, $data) === false</error>,
        <error descr="This comparison is always false: false|int and true share no value.">$c->compare('a', 'b') === true</error>,
    ];
}
