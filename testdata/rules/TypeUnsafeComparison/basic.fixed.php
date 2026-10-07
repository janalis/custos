<?php
interface Labelled            { public function __toString(); }
interface Plain               {}
abstract class BaseTag        { abstract public function __toString(); }
class Tag extends BaseTag     { function __toString(): string { return "t"; } }
class ChildTag extends Tag    {}
class Invoice                 {}
class Stamp extends DateTimeImmutable {}
trait Printable               { function __toString() { return ''; } }
class Note                    { use Printable; }

function compare(Labelled $l, Plain $p, ChildTag $t, Invoice $i, ?Invoice $n, Note $note, $v, $w) {
    return [
        $l == 'x',
        $p != 'x',
        'name' == $t,
        '0' == $i,
        ($n) <> '',
        ($n) <> 'x',
        $note == 'n',

        $v === 'ready',
        "done" !== $v,
        $v!=='1x3',

        $v == '12',
        $v != '-.5',
        $v == "",
        $v <> $w,
        $v == ('text'),

        new Stamp() == new DateTime(),
        date_create('now') != $w,
        $v === 'ready',
    ];
}
