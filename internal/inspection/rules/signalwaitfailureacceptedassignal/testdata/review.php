<?php

function laterStatement($set)
{
    if (<warning descr="Reject signal-wait failure before handling the result.">$sig = pcntl_sigwaitinfo($set)</warning>) {
        echo 1;
        handle($sig);
    }
}

function nestedPlainCondition($set, $verbose)
{
    if (<warning descr="Reject signal-wait failure before handling the result.">$sig = pcntl_sigwaitinfo($set)</warning>) {
        if ($verbose) {
            handle($sig);
        }
    }
}

function guardOnlyCoversOneCall($set)
{
    $sig = pcntl_sigwaitinfo($set);
    if (<warning descr="Reject signal-wait failure before handling the result.">$sig</warning>) {
        record($sig);
        if ($sig > 0) {
            handle($sig);
        }
    }
}

function guardDoesNotLeave($set)
{
    $sig = pcntl_sigwaitinfo($set);
    if ($sig === -1) {
        echo 'failed';
    }
    if (<warning descr="Reject signal-wait failure before handling the result.">$sig</warning>) {
        handle($sig);
    }
}
