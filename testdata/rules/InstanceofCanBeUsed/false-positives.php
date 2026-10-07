<?php

final class Invoice {}
interface Printable {}

function noReport(Invoice $doc, string $name, array $args) {
    return [
        get_class($doc) === $name,          // class not a literal
        $name === get_class($doc),
        get_class($doc, 1) === 'Invoice',   // argument count
        get_class($doc) < 'Invoice',        // not an equality
        is_a('Invoice', 'Invoice'),         // string subject
        is_a(...$args),                     // spread arguments
        is_a($doc, 'Invoice', false, 1),    // argument count
        is_a($doc, $name),
        in_array('Printable'),              // argument count
        in_array('Printable', $args),       // haystack is not a call
        in_array('Printable', array_keys($args)),
        in_array('Printable', $fn($doc)),
        in_array('Printable', class_implements()),
        in_array($name, class_implements($doc)), // needle not a literal
        strlen($name),
        $fn($doc),
    ];
}
