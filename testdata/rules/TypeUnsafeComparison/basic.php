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
        <error descr="\Plain has no __toString(), so it cannot be compared to a string.">$p != 'x'</error>,
        'name' == $t,
        <error descr="\Invoice has no __toString(), so it cannot be compared to a string.">'0' == $i</error>,
        ($n) <> '',
        <error descr="\Invoice has no __toString(), so it cannot be compared to a string.">($n) <> 'x'</error>,
        $note == 'n',

        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$v == 'ready'</warning>,
        <warning descr="Use '!==' here; the string is not numeric, so strict comparison is safe.">"done" != $v</warning>,
        <warning descr="Use '!==' here; the string is not numeric, so strict comparison is safe.">$v<>'1x3'</warning>,

        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == '12'</weak_warning>,
        <weak_warning descr="Prefer '!==' to avoid implicit type juggling.">$v != '-.5'</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == ""</weak_warning>,
        <weak_warning descr="Prefer '!==' to avoid implicit type juggling.">$v <> $w</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == ('text')</weak_warning>,

        new Stamp() == new DateTime(),
        date_create('now') != $w,
        $v === 'ready',
    ];
}
