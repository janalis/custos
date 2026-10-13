<?php
function shapes($v, $w, $x, $y, $o) {
    $a = $v instanceof \Traversable || <weak_warning descr="Use 'is_iterable($v)' instead of the is_array()/instanceof Traversable pair.">is_array($v)</weak_warning>;
    $b = $x || <weak_warning descr="Use 'is_iterable($v)' instead of the is_array()/instanceof Traversable pair.">is_array($v)</weak_warning> || ($v instanceof Traversable);
    $c = $x || $v instanceof Traversable || <weak_warning descr="Use 'is_iterable($v)' instead of the is_array()/instanceof Traversable pair.">is_array($v)</weak_warning>;
    $d = <weak_warning descr="Use 'is_iterable($v)' instead of the is_array()/instanceof Traversable pair.">is_array($v)</weak_warning> || $x || $v instanceof Traversable;         // not neighbours: report only
    $e = $x || <weak_warning descr="Use 'is_iterable($v)' instead of the is_array()/instanceof Traversable pair.">is_array($v)</weak_warning> || $y || $v instanceof Traversable;   // not neighbours: report only
    $f = <weak_warning descr="Use 'is_iterable($v)' instead of the is_array()/instanceof Traversable pair.">is_array($v)</weak_warning> /* why */ || $v instanceof Traversable;     // comment inside the pair: report only
    $g = <weak_warning descr="Use 'is_iterable($v)' instead of the is_array()/instanceof Traversable pair.">is_array($v)</weak_warning> || $v instanceof Iterator || $v instanceof Traversable;
    $h = <weak_warning descr="Use 'is_iterable($v)' instead of the is_array()/instanceof Traversable pair.">is_array($v)</weak_warning> || $w instanceof Traversable || $v instanceof Traversable;
    $i = <weak_warning descr="Use 'is_iterable($$v)' instead of the is_array()/instanceof Traversable pair.">is_array($$v)</weak_warning> || $$v instanceof Traversable;              // variable variable: report only
    $j = <weak_warning descr="Use 'is_iterable($o-&gt;rows())' instead of the is_array()/instanceof Traversable pair.">is_array($o->rows())</weak_warning> || $o->rows() instanceof Traversable;       // ordinary call: report only
    $k = <weak_warning descr="Use 'is_iterable(load($x))' instead of the is_array()/instanceof Traversable pair.">is_array(load($x))</weak_warning> || load($x) instanceof Traversable;
    $l = <weak_warning descr="Use 'is_iterable($v)' instead of the is_array()/instanceof Traversable pair.">\is_array($v,)</weak_warning> || $v instanceof \Traversable;
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k, $l];
}
