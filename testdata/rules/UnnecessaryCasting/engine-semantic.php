<?php
function anonymousMembers() {
    $item = new class {
        public function count(): int { return 1; }
    };
    return <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $item->count();
}
/**
 * @param array $items
 * @psalm-param array<int, int> $items
 * @phpstan-param array<string, int> $items
 */
function documentedKeys(array $items) {
    foreach ($items as $key => $value) {
        echo 'key:' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> $key;
        // Documentation alone cannot justify removing standalone casts.
        echo (string) $key;
        echo (int) $value;
    }
}
