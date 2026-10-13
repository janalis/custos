<?php

function earlierLoop(array $fibers): void
{
    foreach ($fibers as $f) {
        $f->start();
    }
    $f = new Fiber(fn() => 1);
    $f->start();
    <error descr="Resume a fiber before it terminates.">$f->resume()</error>;
}

function callableThenStart(): void
{
    $f = new Fiber(function () { return null; });
    $starter = $f->start(...);
    $f->start();
    <error descr="Resume a fiber before it terminates.">$f->resume()</error>;
}

function otherLoopVariable(array $items): void
{
    $f = new Fiber(fn() => true);
    $f->start();
    foreach ($items as $key => [$item]) {
        echo $key, $item;
    }
    <error descr="Resume a fiber before it terminates.">$f->resume()</error>;
}
