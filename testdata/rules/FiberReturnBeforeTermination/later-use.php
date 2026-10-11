<?php
function readTooEarly() {
    $f = new Fiber(function () { Fiber::suspend(); return 1; });
    $f->start();
    $value = <error descr="Read the fiber return after termination.">$f->getReturn()</error>;
    $f->resume();
    return $value;
}
