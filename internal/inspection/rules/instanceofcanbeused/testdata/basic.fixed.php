<?php

final class Invoice        {}
class Document             {}
final class Receipt extends Document {}
class Draft                {}
interface Printable        {}

function check(Invoice $doc, Document $base, Receipt $rc, Draft $dr, string $name, $loose) {
    return [
        $doc instanceof \Invoice,
        !$rc instanceof \Receipt,
        $doc instanceof \Invoice,
        $base instanceof \Receipt,

        get_class($dr) == 'Draft',
        get_class($doc) === 'invoice',
        get_parent_class($base) == "Document",
        is_subclass_of($base, 'Receipt'),
        in_array('Document', class_parents($base), true),

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
