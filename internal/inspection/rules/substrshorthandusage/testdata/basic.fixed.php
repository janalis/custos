<?php
function cut($name, $head, $k) {
    $a = substr($name, 0, -1);
    $b = mb_substr($name, 3, -2, 'UTF-8');
    $c = \substr($name, 7);
    $d = mb_substr($name, $k);
    $e = mb_substr(
        $name, strlen($head), null, 'UTF-8'
    );
    $f = substr($name, 010, -2);
    $g = substr($this->v, 1);
    return [$a, $b, $c, $d, $e, $f, $g];
}
