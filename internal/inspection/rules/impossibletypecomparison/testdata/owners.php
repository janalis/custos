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

// NotOptimalIfConditions reports a comparison with a literal next to an
// instanceof of the same value in an if condition.
final class Ticket {}
function chains($o): void
{
    if ($o instanceof Ticket && <weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">$o === 'x'</weak_warning>) {
        echo 1;
    }
}

// How NotOptimalIfConditions splits a condition: an elseif, a negated
// chain, a longer chain; a chain with no instanceof, or comparing another
// value, stays this rule's.
function splits($o, int $n): void
{
    if ($n > 1) {
        echo 0;
    } elseif ($o instanceof Ticket && <weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">$o === []</weak_warning>) {
        echo 1;
    }
    if (!($o instanceof Ticket && <weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">$o === -1</weak_warning>)) {
        echo 2;
    }
    if ($n > 1 && $o instanceof Ticket && <weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">"a{$n}" === $o</weak_warning>) {
        echo 3;
    }
    if ($n > 1 && <error descr="This comparison is always false: int and string share no value.">$n === 'x'</error>) {
        echo 4;
    }
    if ($o instanceof Ticket && <error descr="This comparison is always false: int and string share no value.">$n === 'x'</error>) {
        echo 5;
    }
    // An operand under any prefix operator counts as a literal there.
    if ($o instanceof Ticket && <weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">$o === !false</weak_warning>) {
        echo 6;
    }
}
