<?php

final class Invoice implements Countable { function count(): int { return 0; } }
class Document {}

function tally(Invoice $doc, ?Invoice $primary, Document $fallback, bool $flag) {
    $kept = null;
    return [
        (int) ($doc instanceof \Invoice),
        (string) ($doc instanceof \Countable),
        -($doc instanceof \Invoice),
        ($doc instanceof \Invoice) ** 2,
        !$doc instanceof \Invoice,
        1 + $doc instanceof \Invoice,
        ($primary ?: $fallback) instanceof \Invoice,
        ($primary ?? $fallback) instanceof \Invoice,
        ($flag ? $fallback : $doc) instanceof \Invoice,
        ($kept = $doc) instanceof \Invoice,
        (!$flag) instanceof \Invoice,
        !($kept = $doc) instanceof \Invoice,
        is_subclass_of($primary ?: $fallback, 'Document'),
        $kept,
    ];
}
