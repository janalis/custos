<?php

function startedAgainAfterCallable(): void
{
    $job = new Fiber(fn() => Fiber::suspend());
    $job->start();
    $launch = $job->start(...);
    $other = [1];
    foreach ($other as $key => $value) {
        echo $key, $value;
    }
    <error descr="Start each fiber only once.">$job->start()</error>;
}
