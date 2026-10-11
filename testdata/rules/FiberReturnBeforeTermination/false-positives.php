<?php
function drainLoop() {
    $f = new Fiber(function () { Fiber::suspend(); return 1; });
    $f->start();
    while ($f->isSuspended()) { $f->resume(); }
    return $f->getReturn();
}
function resumeWhenPending() {
    $f = new Fiber(function () { Fiber::suspend(); return 2; });
    $f->start();
    if (!$f->isTerminated()) { $f->resume(); }
    return $f->isTerminated() ? $f->getReturn() : null;
}
function resumeInTry() {
    $f = new Fiber(function () { Fiber::suspend(); return 3; });
    $f->start();
    try { $f->resume(); } finally { }
    return $f->getReturn();
}
function onlyWhenDone() {
    $f = new Fiber(function () { Fiber::suspend(); return 4; });
    $f->start();
    if ($f->isTerminated()) { return $f->getReturn(); }
    return 'pending';
}
function drivenElsewhere() {
    $f = new Fiber(function () { Fiber::suspend(); return 5; });
    $f->start();
    $step = fn() => $f->resume();
    $step();
    return $f->getReturn();
}
function drivenByLaterIteration() {
    $f = new Fiber(function () { Fiber::suspend(); return 6; });
    $f->start();
    for ($i = 0; $i < 2; $i++) {
        if ($i > 0) { return $f->getReturn(); }
        $f->resume();
    }
}
echo json_encode([drainLoop(), resumeWhenPending(), resumeInTry(), onlyWhenDone(), drivenElsewhere(), drivenByLaterIteration()]), "\n";
