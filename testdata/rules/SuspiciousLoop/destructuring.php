<?php
function pairs(array $list, $a) {
    <error descr="Loop variable '$a' overwrites a function parameter.">foreach</error> ($list as [$a, $b]) {
        <error descr="Loop variable '$b' overwrites a variable of an outer loop.">foreach</error> ($list as $k => &$b) {}
    }
}

function legacy(array $rows, $c) {
    <error descr="Loop variable '$c' overwrites a function parameter.">foreach</error> ($rows as list($c, , list($d))) {}
}
