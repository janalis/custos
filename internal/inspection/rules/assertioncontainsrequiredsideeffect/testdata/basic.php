<?php
<warning descr="Move required assignments outside assertions.">assert(($result = 7) > 0)</warning>; echo $result;
function inOutArgument(): array {
    <warning descr="Move required assignments outside assertions.">assert(($pair = [3, 1]) !== [])</warning>;
    sort($pair);
    return $pair;
}
function appendedOutput(): array {
    <warning descr="Move required assignments outside assertions.">assert(($lines = ['first']) !== [])</warning>;
    exec('true', $lines);
    return $lines;
}
function iteratesItself(): int {
    <warning descr="Move required assignments outside assertions.">assert(($items = [1, 2]) !== [])</warning>;
    foreach ($items as $items) {
    }
    return 0;
}
function readBeforeLoop(array $list): int {
    <warning descr="Move required assignments outside assertions.">assert(($seen = count($list)) >= 0)</warning>;
    $total = $seen;
    foreach ($list as $seen) {
        $total += $seen;
    }
    return $total;
}
function patternAndMatches(string $text): int {
    <warning descr="Move required assignments outside assertions.">assert(($hit = '/a/') !== '')</warning>;
    return preg_match($hit, $text, $hit);
}
function namedOutArgument(string $text): array {
    <warning descr="Move required assignments outside assertions.">assert(($found = []) === [])</warning>;
    preg_match('/a/', $text, matches: $found);
    return $found;
}
function readAfterSecondAssert(int $n): int {
    <warning descr="Move required assignments outside assertions.">assert(($half = intdiv($n, 2)) >= 0)</warning>;
    assert($half <= $n);
    return $half;
}
function userLookalike(string $s): array {
    <warning descr="Move required assignments outside assertions.">assert(($m = []) === [])</warning>;
    Custom\preg_match('/\d+/', $s, $m);
    return [];
}
