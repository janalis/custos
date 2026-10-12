<?php
// A guard covers the result it tested, not one assigned afterwards.
function loopReassigns($first, $next, $count)
{
    $parts = parse_url($first);
    if ($parts === false) {
        return;
    }
    while ($count--) {
        echo <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
        $parts = parse_url($next);
    }
}

function onePathReassigns($first, $next, $again)
{
    $parts = parse_url($first);
    if ($parts === false) {
        return;
    }
    if ($again) {
        $parts = parse_url($next);
    }
    return <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
}

function literalBeforeLoop($next, $count)
{
    $parts = ['host' => 'localhost'];
    foreach ($count as $unused) {
        echo <warning descr="Reject URL parsing failure before indexing.">$parts['host']</warning>;
        $parts = parse_url($next);
    }
}

// Tested where every assignment has been made.
function testedInsideLoop($first, $next, $count)
{
    $parts = parse_url($first);
    while ($count--) {
        if ($parts === false) {
            return;
        }
        echo $parts['host'];
        $parts = parse_url($next);
    }
}

function keyTestedInsideLoop($first, $next, $count)
{
    $parts = parse_url($first);
    while ($count--) {
        if (isset($parts['host'])) {
            echo $parts['host'];
        }
        $parts = parse_url($next);
    }
}

function testedAgain($first, $next, $again)
{
    $parts = parse_url($first);
    if ($again) {
        $parts = parse_url($next);
    }
    if (!$parts) {
        return;
    }
    return $parts['host'];
}

// The loop replaces the tested result with something that is not a URL parse.
function loopClears($first, $count)
{
    $parts = parse_url($first);
    if ($parts === false) {
        return;
    }
    while ($count--) {
        echo $parts['host'];
        $parts = load($count);
    }
}
