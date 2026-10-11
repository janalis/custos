<?php

function loopVariable(): mixed
{
    $f = new Fiber(fn() => 1);
    $f->start();
    foreach ([new Fiber(function () { return Fiber::suspend('s'); })] as $f) {
        $f->start();
        $f->resume('r');
    }
    return $f->getReturn();
}

function destructured(): void
{
    $f = new Fiber(fn() => 1);
    $f->start();
    [$f] = [new Fiber(function () { Fiber::suspend(); })];
    $f->start();
    $f->resume();
}

function callables(): bool
{
    $f = new Fiber(fn() => 1);
    $starter = $f->start(...);
    $resumer = $f->resume(...);
    $f->resume(...);
    return $f->isStarted();
}

function startCallableOnly(): void
{
    $f = new Fiber(fn() => 1);
    $starter = $f->start(...);
    $f->resume();
}

function referenced(): void
{
    $f = new Fiber(fn() => 1);
    $f->start();
    $swap = function () use (&$f) {
        $f = new Fiber(function () { Fiber::suspend(); });
        $f->start();
    };
    $swap();
    $f->resume();
}
