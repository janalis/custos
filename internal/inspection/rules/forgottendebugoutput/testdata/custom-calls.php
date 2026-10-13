<?php
namespace Acme;

class Probe {
    public function trace($v) { return $v; }
    public function other($v) { return $v; }
}

function run(Probe $p, $m) {
    <error descr="Debug output call; remove it if it was left over from debugging.">my_trace(1)</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">$p->trace(1)</error>;
    $p->other(1);
    $p->$m(1);
    unrelated(1);
}
