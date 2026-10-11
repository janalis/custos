<?php
<error descr="Supply a nonnegative repetition count.">str_repeat('cd', -3)</error>;

function countdown(): string
{
    $pad = -1;
    $pad--;
    return str_repeat('-', $pad);
}

function steppedLater(): string
{
    $pad = -1;
    $out = <error descr="Supply a nonnegative repetition count.">str_repeat('-', $pad)</error>;
    $pad++;
    return $out;
}

function arrowCopy(): string
{
    $pad = -1;
    $next = fn() => $pad++;
    return <error descr="Supply a nonnegative repetition count.">str_repeat('-', $pad)</error> . $next();
}

function lateCapture(): string
{
    $pad = -1;
    $out = <error descr="Supply a nonnegative repetition count.">str_repeat('-', ($pad))</error>;
    $set = function () use (&$pad, $out) { $pad = 2; };
    $set();
    return $out;
}

function valueCapture(): string
{
    $pad = -1;
    $show = function () use ($pad) { return $pad + 1; };
    return <error descr="Supply a nonnegative repetition count.">str_repeat('-', $pad)</error> . $show();
}

function copiedBeforeStep(): string
{
    $n = -1;
    $m = $n;
    $n++;
    return <error descr="Supply a nonnegative repetition count.">str_repeat('a', $m)</error>;
}

function reboundLater(array $vars): string
{
    $n = -1;
    $out = <error descr="Supply a nonnegative repetition count.">str_repeat('a', $n)</error>;
    extract($vars);
    return $out;
}
