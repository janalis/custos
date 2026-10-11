<?php
function onlyAsserts(int $n): string {
    assert(($half = intdiv($n, 2)) >= 0);
    assert($half <= $n);
    return "ok";
}
function outParam(string $s): array {
    assert(($m = []) === []);
    preg_match('/\d+/', $s, $m);
    return $m;
}
function secondOutParam(): bool {
    assert(($line = 0) === 0);
    $sent = headers_sent($file, $line);
    return $sent && $line > 0;
}
function foreachTarget(array $list): int {
    assert(($v = count($list)) >= 0);
    $sum = 0;
    foreach ($list as $v) {
        $sum += $v;
    }
    return $sum;
}
function foreachKey(array $list): string {
    assert(($k = 'none') !== '');
    $out = '';
    foreach ($list as $k => $val) {
        $out .= $k;
    }
    return $out;
}
function foreachDestructuring(array $rows): int {
    assert(($left = 0) === 0);
    $sum = 0;
    foreach ($rows as [$left, $right]) {
        $sum += $left + $right;
    }
    return $sum;
}
function catchVariable(callable $task): string {
    assert(($e = null) === null);
    try {
        $task();
    } catch (Exception $e) {
        return $e->getMessage();
    }
    return '';
}
