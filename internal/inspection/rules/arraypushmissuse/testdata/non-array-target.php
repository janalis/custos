<?php
function fill(ArrayAccess $bag, $unknown, ?array $maybe, string $text, $job) {
    array_push($bag, $job);
    array_push($unknown, $job);
    array_push($maybe, $job);
    array_push($text, $job);
    array_push($this->items, $job);
}
