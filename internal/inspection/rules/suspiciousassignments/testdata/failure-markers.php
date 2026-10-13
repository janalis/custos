<?php
/** @return int[]|false */
function sample_pair() { return [1, 2]; }

/** @return int[]|int|false */
function mixed_pair() { return 0; }

/** @return array|bool */
function validate_filter(string $f) { return $f === '' ? false : explode('-', $f); }

function bool_markers(string $f, bool $flag) {
    $filter = validate_filter($f);
    if (!$filter) {
        return;
    }
    [$from, $to] = $filter;              // custos: true/bool next to array
    <error descr="Destructuring a value that is not an array.">[$v] = $flag</error>;   // only bool
}

function failure_markers(array|false $found, ?array $cached, string|false $line, false $none,
                         array|null|false $both, \ArrayObject|false $obj, array|string $mixedUp, ?string $text) {
    [$p, $q] = $found;                   // E8: false ignored next to array
    [$r] = $cached;                      // E8: null ignored next to array
    [$s, $t] = sample_pair();            // E8: int[]|false
    [$a, $b] = $both;                    // E8: array|null|false
    [$c] = $obj;                         // E8: ArrayAccess|false
    <error descr="Destructuring a value that is not an array.">[$u] = $line</error>;   // string remains
    <error descr="Destructuring a value that is not an array.">[$w] = $none</error>;   // only false
    <error descr="Destructuring a value that is not an array.">[$x] = $mixedUp</error>;  // string next to array
    <error descr="Destructuring a value that is not an array.">[$y] = mixed_pair()</error>;  // int remains
    <error descr="Destructuring a value that is not an array.">list($z) = $text</error>;  // no supporting type
}

function colour(string $c) {
    list($r, $g, $b) = sscanf($c, "#%02x%02x%02x");   // array|null without output variables
    return $r + $g + $b;
}
