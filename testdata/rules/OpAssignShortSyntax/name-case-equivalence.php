<?php
function grow($o, $n) {
    <weak_warning descr="Use the compound form '$o-&gt;sums[slot::ID()] *= $n'.">$o->sums[Slot::Id()] = $o->sums[slot::ID()] * $n</weak_warning>;
    $o->Total = $o->total * $n;
}
