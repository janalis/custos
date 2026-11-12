<?php
// Results other rules report compared with a constant they never equal:
// left to them when they run. ArraySearchUsedAsInArray reads the call only.
function owners(array $roles, string $who, $parser, string $data, Collator $c): array
{
    $found = array_search($who, $roles, true);
    return [
        array_search($who, $roles, true) === <error descr="array_search() cannot return true; this comparison never changes.">true</error>,
        <error descr="array_search() cannot return true; this comparison never changes.">true</error> !== array_search($who, $roles, true),
        <error descr="This comparison is always true: false|int|string and true share no value.">$found !== true</error>,
        <warning descr="Compare XML parsing failure with integer zero.">xml_parse($parser, $data) === false</warning>,
        <warning descr="Compare collation results explicitly.">$c->compare('a', 'b') === true</warning>,
    ];
}
