<?php
/**
 * @template K of array-key
 * @template V
 * @param array<K, V> $items
 * @return K
 */
function templateKey(array $items) {
    foreach ($items as $key => $value) {
        return $key;
    }
    throw new RuntimeException('An entry is required');
}

/** @param array<string, int> $items */
function templateKeyLabels(array $items) {
    echo 'key:' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> templateKey($items);
    echo 'named:' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> templateKey(items: ['first' => 1]);
    // A documented contract alone cannot justify a standalone cast fix.
    echo (string) templateKey($items);
}

function uncertainTemplateKey(array $items) {
    echo 'unknown:' . (string) templateKey($items);
    echo 'empty:' . (string) templateKey([]);
}
