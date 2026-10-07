<?php
class Stats { public static $n = 0; public static $v = 0; }
function tally() {
    <error descr="The target is repeated on the right-hand side of the compound assignment; likely a merge mistake.">Stats::$n += stats::$n + 1</error>;
    Stats::$v = 1;
    <error descr="STATS::$v is overwritten right after being assigned.">STATS::$v</error> = 2;
}
