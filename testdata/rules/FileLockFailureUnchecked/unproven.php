<?php

function reassignedResult($h)
{
    $ok = flock($h, LOCK_EX);
    $ok = true;
    if (!$ok) {
        return;
    }
    <warning descr="Acquire the lock successfully before writing.">fwrite($h, 'entry')</warning>;
}

function testedOnOnePath($h, $verbose)
{
    $ok = flock($h, LOCK_EX);
    if ($verbose) {
        if (!$ok) {
            return;
        }
    }
    <warning descr="Acquire the lock successfully before writing.">fwrite($h, 'entry')</warning>;
}

function failureBranchContinues($h)
{
    $ok = flock($h, LOCK_EX);
    if (!$ok) {
        error_log('lock failed');
    }
    <warning descr="Acquire the lock successfully before writing.">fwrite($h, 'entry')</warning>;
}

function writesOnFailure($h, $g, $i)
{
    if (flock($h, LOCK_EX) === false) {
        <warning descr="Acquire the lock successfully before writing.">fwrite($h, 'entry')</warning>;
    }
    if (flock($g, LOCK_EX) === true) {
        return;
    } else {
        <warning descr="Acquire the lock successfully before writing.">fwrite($g, 'entry')</warning>;
    }
    $ok = flock($i, LOCK_EX);
    return $ok ? false : <warning descr="Acquire the lock successfully before writing.">fwrite($i, 'entry')</warning>;
}

function unrelatedTests($h, $g, $limit)
{
    if (flock($h, LOCK_EX) === $limit) {
        return;
    }
    <warning descr="Acquire the lock successfully before writing.">fwrite($h, 'entry')</warning>;
    $count = (int) flock($g, LOCK_EX);
    <warning descr="Acquire the lock successfully before writing.">fwrite($g, 'entry')</warning>;
    return $count;
}

function weakShortCircuit($h, $g, $i, $force)
{
    flock($h, LOCK_EX) || <warning descr="Acquire the lock successfully before writing.">fwrite($h, 'entry')</warning>;
    flock($g, LOCK_EX) || error_log('lock failed');
    <warning descr="Acquire the lock successfully before writing.">fwrite($g, 'entry')</warning>;
    if (flock($i, LOCK_EX) === true || $force) {
        <warning descr="Acquire the lock successfully before writing.">fwrite($i, 'entry')</warning>;
    }
}

function weakConjunction($h, $g, $force)
{
    if (flock($h, LOCK_EX) === false && $force) {
        return;
    }
    <warning descr="Acquire the lock successfully before writing.">fwrite($h, 'entry')</warning>;
    $ok = flock($g, LOCK_EX);
    $ok ?: <warning descr="Acquire the lock successfully before writing.">fwrite($g, 'entry')</warning>;
}

function retryCanGiveUp($h, $g, $tries)
{
    do {
        $ok = flock($h, LOCK_EX | LOCK_NB);
        if (--$tries === 0) {
            break;
        }
    } while (!$ok);
    <warning descr="Acquire the lock successfully before writing.">fwrite($h, 'entry')</warning>;
    while (flock($g, LOCK_EX | LOCK_NB) === true) {
        usleep(10);
    }
    <warning descr="Acquire the lock successfully before writing.">fwrite($g, 'entry')</warning>;
}

function testedInOtherScope($h)
{
    $ok = flock($h, LOCK_EX);
    $check = function () use ($ok) {
        if (!$ok) {
            return;
        }
    };
    $check();
    <warning descr="Acquire the lock successfully before writing.">fwrite($h, 'entry')</warning>;
}
