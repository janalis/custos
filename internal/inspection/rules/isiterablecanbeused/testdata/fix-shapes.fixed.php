<?php
function shapes($v, $w, $x, $y, $o) {
    $a = is_iterable($v);
    $b = $x || is_iterable($v);
    $c = $x || is_iterable($v);
    $d = is_array($v) || $x || $v instanceof Traversable;         // not neighbours: report only
    $e = $x || is_array($v) || $y || $v instanceof Traversable;   // not neighbours: report only
    $f = is_array($v) /* why */ || $v instanceof Traversable;     // comment inside the pair: report only
    $g = is_array($v) || $v instanceof Iterator || $v instanceof Traversable;
    $h = is_array($v) || $w instanceof Traversable || $v instanceof Traversable;
    $i = is_array($$v) || $$v instanceof Traversable;              // variable variable: report only
    $j = is_array($o->rows()) || $o->rows() instanceof Traversable;       // ordinary call: report only
    $k = is_array(load($x)) || load($x) instanceof Traversable;
    $l = \is_iterable($v);
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k, $l];
}
