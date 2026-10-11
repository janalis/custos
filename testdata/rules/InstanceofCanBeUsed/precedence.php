<?php

final class Invoice implements Countable { function count(): int { return 0; } }
class Document {}

function tally(Invoice $doc, ?Invoice $primary, Document $fallback, bool $flag) {
    $kept = null;
    return [
        (int) <warning descr="Prefer '$doc instanceof \Invoice'.">is_a($doc, 'Invoice')</warning>,
        (string) <warning descr="Prefer '$doc instanceof \Countable'.">in_array('Countable', class_implements($doc))</warning>,
        -<warning descr="Prefer '$doc instanceof \Invoice'.">is_a($doc, 'Invoice')</warning>,
        <warning descr="Prefer '$doc instanceof \Invoice'.">is_a($doc, 'Invoice')</warning> ** 2,
        !<warning descr="Prefer '$doc instanceof \Invoice'.">is_a($doc, 'Invoice')</warning>,
        1 + <warning descr="Prefer '$doc instanceof \Invoice'.">is_a($doc, 'Invoice')</warning>,
        <warning descr="Prefer '($primary ?: $fallback) instanceof \Invoice'.">is_a($primary ?: $fallback, 'Invoice')</warning>,
        <warning descr="Prefer '($primary ?? $fallback) instanceof \Invoice'.">is_a($primary ?? $fallback, 'Invoice')</warning>,
        <warning descr="Prefer '($flag ? $fallback : $doc) instanceof \Invoice'.">is_a($flag ? $fallback : $doc, 'Invoice')</warning>,
        <warning descr="Prefer '($kept = $doc) instanceof \Invoice'.">is_a($kept = $doc, 'Invoice')</warning>,
        <warning descr="Prefer '(!$flag) instanceof \Invoice'.">is_a(!$flag, 'Invoice')</warning>,
        <warning descr="Prefer '!($kept = $doc) instanceof \Invoice'.">get_class($kept = $doc) !== 'Invoice'</warning>,
        <warning descr="Consider '($primary ?: $fallback) instanceof \Document' (not an exact equivalent).">is_subclass_of($primary ?: $fallback, 'Document')</warning>,
        $kept,
    ];
}
