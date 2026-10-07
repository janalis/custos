<?php
class Counter { public static $hits = 0; const FIRST = 0; }
function bump(array $counts) {
    <weak_warning descr="Use '++Counter::$hits' instead.">Counter::$hits = counter::$hits + 1</weak_warning>;
    <weak_warning descr="Use '--$counts[Counter::FIRST]' instead.">$counts[Counter::FIRST] = $counts[COUNTER::FIRST] - 1</weak_warning>;
}
