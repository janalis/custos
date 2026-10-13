<?php
function anonymousMembers() {
    $item = new class {
        public function count(): int { return 1; }
    };
    return $item->count();
}
/**
 * @param array $items
 * @psalm-param array<int, int> $items
 * @phpstan-param array<string, int> $items
 */
function documentedKeys(array $items) {
    foreach ($items as $key => $value) {
        echo 'key:' . $key;
        // Documentation alone cannot justify removing standalone casts.
        echo (string) $key;
        echo (int) $value;
    }
}
