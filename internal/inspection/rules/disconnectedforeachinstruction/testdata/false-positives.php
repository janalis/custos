<?php
foreach ($orders as $order) {
    $box = new \ArrayObject();
    $count++;
    $last = $order;
    if ($limitReached) { break; }
    $list[] = $order;
}
foreach ($items as [$a, $b]) {
    echo $a;
    work($b);
}
function f(array $stack) {
    foreach ($stack as $x) {
        array_pop($stack);
        echo $stack;
    }
}
foreach ($rows as $row) echo $title;
final class Widths
{
    private static $table;
    public static function measure(array $codes) {
        foreach ($codes as $code) {
            self::$table ??= load_table();
            static::$table = [];
            echo $code;
        }
    }
}
function nested(array $rows) {
    foreach ($rows as $row) {
        [$id, [$first, $last]] = $row;
        render($last);
        ['meta' => ['tag' => $tag]] = $row;
        render($tag);
        list($a, list($b, list($c))) = $row;
        render($c);
        foreach ($row as [$k, [$deep]]) {
            touch($k);
        }
        render($deep);
    }
}
function joined(array $rows, string $sep, array $head) {
    foreach ($rows as $row) {
        yield $sep;
        yield from $head;
        if ($sep) {
            yield $sep => $head;
        }
        yield $row;
    }
}
