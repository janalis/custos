<?php
class Counter { public static $hits = 0; const FIRST = 0; }
function bump(array $counts) {
    ++Counter::$hits;
    --$counts[Counter::FIRST];
}
