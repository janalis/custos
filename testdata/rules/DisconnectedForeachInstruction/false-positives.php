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
