<?php
function setup() {
    <weak_warning descr="Split this chained assignment into separate assignments.">$width = $height = 64</weak_warning>;
    <weak_warning descr="Split this chained assignment into separate assignments.">$first = $last = "none"</weak_warning>;
    <weak_warning descr="Split this chained assignment into separate assignments.">$p = $q = $r = make_point(1, 2)</weak_warning>;
    <weak_warning descr="Split this chained assignment into separate assignments.">$left = $right = $origin</weak_warning>;
    <weak_warning descr="Split this chained assignment into separate assignments.">$lo = $hi = -1</weak_warning>;
    <weak_warning descr="Split this chained assignment into separate assignments.">$s = $t = "x$y"</weak_warning>;
    <weak_warning descr="Split this chained assignment into separate assignments.">$c = $d = Mode::FAST</weak_warning>;

    if (<weak_warning descr="Split this chained assignment into separate assignments.">$m = $n = fetch()</weak_warning>) {}
    <weak_warning descr="Split this chained assignment into separate assignments.">$u = $v += 1</weak_warning>;
    <weak_warning descr="Split this chained assignment into separate assignments.">$w = $z = &$ref</weak_warning>;
    return <weak_warning descr="Split this chained assignment into separate assignments.">$e = $f = 1</weak_warning>;
}
