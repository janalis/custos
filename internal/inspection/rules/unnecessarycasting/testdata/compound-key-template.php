<?php
/**
 * @template K
 * @param array<K|int, int> $items
 * @return K
 */
function selectedKey($items) {
    return array_key_first($items);
}

function keyLabels(array $unknown) {
    return [
        'key=' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> selectedKey(['name' => 1]),
        'key=' . (string) selectedKey([1]),
        'key=' . (string) selectedKey($unknown),
    ];
}
