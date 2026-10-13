<?php

function report(): void
{
    error_log('unreadable');
}

function alternativeReturns($path)
{
    $text = file_get_contents($path) or report();

    return <warning descr="Handle file reading failure before using its result.">strlen($text)</warning>;
}

function branchReturns($path)
{
    $text = file_get_contents($path);
    if ($text === false) {
        report();
    }

    return <warning descr="Handle file reading failure before using its result.">strlen($text)</warning>;
}

function unrelatedComparison($path)
{
    $text = file_get_contents($path);
    if ($text == 'skip') {
        return 0;
    }

    return <warning descr="Handle file reading failure before using its result.">strlen($text)</warning>;
}

function wrongBranch($path)
{
    $text = file_get_contents($path);

    return $text == false ? <warning descr="Handle file reading failure before using its result.">strlen($text)</warning> : 0;
}

function readAfterGuard($path)
{
    $text = file_get_contents($path);
    if ($text == false) {
        return 0;
    }
    $text = file_get_contents($path . '.bak');

    return <warning descr="Handle file reading failure before using its result.">strlen($text)</warning>;
}
