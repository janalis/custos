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
