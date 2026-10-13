<?php
<error descr="Only the last expression of the 'for' condition is evaluated as the condition; combine them with && or ||.">for</error> ($n = 0; $n < 5, $n != 3; $n++) {}

function scan(array $rows, $pos) {
    <error descr="Loop variable '$rows' overwrites a function parameter.">foreach</error> ($rows as $rows) {}
    <error descr="Loop variable '$pos' overwrites a function parameter.">for</error> ($pos = 1; $pos < 3; $pos++) {}
    foreach ($rows as $r) {}
}

class Grid {
    public function walk($cell) {
        <error descr="Loop variable '$cell' overwrites a method parameter.">foreach</error> ([1, 2] as $cell) {}
    }
}

foreach ($matrix as $row => $cols) {
    <error descr="Loop variable '$row' overwrites a variable of an outer loop.">for</error> ($row = 0, $k = 1; $row < 2; $row++) {}
    while ($cols) {
        <error descr="Loop variable '$cols' overwrites a variable of an outer loop.">foreach</error> ($cols as $cols) {}
    }
    array_map(function () {
        foreach ([] as $row) {}
    }, []);
    foreach ($cols as $c) {}
}
