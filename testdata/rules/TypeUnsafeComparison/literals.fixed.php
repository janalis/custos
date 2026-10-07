<?php
class Ghostly extends Missing {}

/**
 * @param int[] $ids
 * @param Invoice|int $mixed
 */
function literals($v, array $ids, $mixed, Unknown $u) {
    return [
        $v == 1,
        2 == $v,
        $v == `ls`,
        $v === b'abc',
        $v == B"12",
        $v === <<<'TXT'
            done
            TXT,
        $v !== <<<TXT
            ready
            TXT,
        $v == <<<TXT
            tab\there
            TXT,
        $v == <<<TXT
            TXT,
        $ids === 'x',
        $mixed === 'x',
        $u == 'x',
        (new Ghostly()) == 'x', // an unresolved ancestor may declare __toString()
    ];
}

class Invoice {}
