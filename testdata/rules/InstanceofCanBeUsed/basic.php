<?php

final class Invoice        {}
class Document             {}
final class Receipt extends Document {}
class Draft                {}
interface Printable        {}

function check(Invoice $doc, Document $base, Receipt $rc, Draft $dr, string $name, $loose) {
    return [
        <warning descr="Prefer '$doc instanceof \Invoice'.">'Invoice' === get_class($doc)</warning>,
        <warning descr="Prefer '!$rc instanceof \Receipt'.">get_class($rc) <> 'Receipt'</warning>,
        <warning descr="Prefer '$doc instanceof \Invoice'.">\is_a($doc, 'Invoice', FALSE)</warning>,
        <warning descr="Prefer '$base instanceof \Receipt'.">is_a($base, 'Receipt')</warning>,

        <warning descr="Consider '$dr instanceof \Draft' (not an exact equivalent).">get_class($dr) == 'Draft'</warning>,
        <warning descr="Consider '$doc instanceof \invoice' (not an exact equivalent).">get_class($doc) === 'invoice'</warning>,
        <warning descr="Consider '$base instanceof \Document' (not an exact equivalent).">get_parent_class($base) == "Document"</warning>,
        <warning descr="Consider '$base instanceof \Receipt' (not an exact equivalent).">is_subclass_of($base, 'Receipt')</warning>,
        <warning descr="Consider '$base instanceof \Document' (not an exact equivalent).">in_array('Document', class_parents($base), true)</warning>,

        get_class($base) === 'Document',
        (get_class($doc)) === 'Invoice',
        get_class($name) == 'Invoice',
        get_class($loose) == 'Invoice',
        is_a($doc, 'Invoice', true),
        is_a($doc, 'Foo'),
        is_a($doc, '\\Invoice'),
        is_a($doc, 'Printable'),
        in_array('Invoice', class_implements($name)),
    ];
}
