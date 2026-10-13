<?php
function grow($o, $n) {
    $o->sums[slot::ID()] *= $n;
    $o->Total = $o->total * $n;
}
