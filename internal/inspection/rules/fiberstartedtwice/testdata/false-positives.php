<?php

function callableOnly(): mixed
{
    $job = new Fiber(fn() => 1);
    $launch = $job->start(...);
    $job->start();
    return $job->getReturn();
}

function callableAfterStart(): Closure
{
    $job = new Fiber(fn() => 1);
    $job->start();
    return $job->start(...);
}

function destructured(): mixed
{
    $job = new Fiber(fn() => 1);
    $job->start();
    [$job] = [new Fiber(fn() => 2)];
    $job->start();
    return $job->getReturn();
}

function listDestructured(): mixed
{
    $job = new Fiber(fn() => 1);
    $job->start();
    list('next' => $job) = ['next' => new Fiber(fn() => 2)];
    $job->start();
    return $job->getReturn();
}

function loopValue(): mixed
{
    $job = new Fiber(fn() => 1);
    $job->start();
    foreach ([new Fiber(fn() => 3)] as $job) {
        $job->start();
    }
    return $job->getReturn();
}

function caught(): void
{
    $job = new Fiber(fn() => 1);
    $job->start();
    try {
        work();
    } catch (Fiber $job) {
    }
    $job->start();
}
