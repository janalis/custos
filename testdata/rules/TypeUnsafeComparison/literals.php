<?php
class Ghostly extends Missing {}

/**
 * @param int[] $ids
 * @param Invoice|int $mixed
 */
function literals($v, array $ids, $mixed, Unknown $u) {
    return [
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == 1</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">2 == $v</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == `ls`</weak_warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$v == b'abc'</warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == B"12"</weak_warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$v == <<<'TXT'
            done
            TXT</warning>,
        <warning descr="Use '!==' here; the string is not numeric, so strict comparison is safe.">$v != <<<TXT
            ready
            TXT</warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == <<<TXT
            tab\there
            TXT</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == <<<TXT
            TXT</weak_warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$ids == 'x'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$mixed == 'x'</warning>,
        $u == 'x',
        (new Ghostly()) == 'x', // an unresolved ancestor may declare __toString()
    ];
}

class Invoice {}
