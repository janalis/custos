<?php
function demo(array $rows) {
    foreach ($rows as $k => $v) {
        echo $k[0];
    }
    $tail = substr('abc', 1);
    echo $tail[0];
    $obj = new ArrayObject();
    echo $obj[new DateTime()];
    $parts = explode(',', 'a,b');
    echo $parts[0];
    echo $rows[$unknown];
}

/**
 * @param mixed $m
 * @param \Countable|array $c
 * @param null|array $n
 */
function ok($m, $c, $n, $u) {
    return [$m['a'], $c['a'], $n['a'], $u['a']];
}

function unresolved() {
    // The class is not known (missing include, wrong namespace): no report.
    $chart = new MissingChart();
    $chart[0] = 1;
    return $chart['series'];
}

if (!function_exists('mb_strtolower')) {
    // A polyfill: the builtin (string result) is what runs.
    function mb_strtolower($string, $encoding = null) { return polyfill_lower($string); }
}

function polyfilled(string $code) {
    $map = [];
    $map[preg_replace('/-/', '', mb_strtolower($code))] = 1;
    return $map;
}
