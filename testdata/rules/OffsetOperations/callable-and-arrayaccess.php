<?php
function measuredLength(string $text): int { return strlen($text); }

final class CountStore implements ArrayAccess {
    public function offsetExists(mixed $offset): bool { return true; }
    public function offsetGet(mixed $offset): int { return 7; }
    public function offsetSet(mixed $offset, mixed $value): void {}
    public function offsetUnset(mixed $offset): void {}
}

final class RowStore implements ArrayAccess {
    public function offsetExists(mixed $offset): bool { return true; }
    public function offsetGet(mixed $offset): array { return ['title' => 'sample']; }
    public function offsetSet(mixed $offset, mixed $value): void {}
    public function offsetUnset(mixed $offset): void {}
}

function inspectRetrievedValues($unknownCallback, $unknownStore) {
    $measure = measuredLength(...);
    $size = $measure('sample');
    <error descr="'$size' does not support offset access (types: int).">$size[0]</error>;

    $builtin = strlen(...);
    $builtinSize = $builtin('sample');
    <error descr="'$builtinSize' does not support offset access (types: int).">$builtinSize[0]</error>;

    $counts = new CountStore();
    $count = $counts['total'];
    <error descr="'$count' does not support offset access (types: int).">$count[0]</error>;

    $rows = new RowStore();
    $row = $rows['first'];
    echo $row['title'];

    $unknown = $unknownCallback();
    echo $unknown[0];
    $unknownEntry = $unknownStore['first'];
    echo $unknownEntry[0];
}
