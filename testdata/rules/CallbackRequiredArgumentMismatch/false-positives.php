<?php
// One signature per branch: the declaration bound at run time is unknown.
if (PHP_VERSION_ID < 80000) {
    function label($item, $prefix) { return $prefix . $item; }
} else {
    function label($item) { return (string) $item; }
}
array_map('label', $items);

// A guarded fallback may never be the declaration in effect.
if (!function_exists('pair_up')) {
    function pair_up($left, $right) { return [$left, $right]; }
}
array_map('pair_up', $items);

// Declared only when the enclosing function runs.
function install_helpers() {
    function nested_pair($left, $right) { return [$left, $right]; }
}
array_map('nested_pair', $items);

// Enough arrays for the unique declaration.
function sum_pair($left, $right) { return $left + $right; }
array_map('sum_pair', $items, $others);
